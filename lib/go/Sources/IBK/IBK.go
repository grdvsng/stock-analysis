package IBK

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	_ "time/tzdata"

	"github.com/ThomasMarcelis/ibkr-go/v2"
	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Result"
	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Time/Schedule"
)

const PAUSE_STEP = 250 * time.Millisecond

type source struct {
	wg         *sync.WaitGroup
	stop       *atomic.Bool
	mutex      *sync.RWMutex
	connection *Connection

	subscribtions    *Sources.Subscrubtions
	subscrubtionStop map[string]*atomic.Bool
}

func New(host string, port, timeout, clientId int) Sources.Source {
	return Result.Ok(Connect(host, port, timeout, clientId))
}

func Connect(host string, port, timeout, clientId int) (Sources.Source, error) {
	connection, e := Dial(host, port, timeout, clientId)

	if e != nil {
		return nil, e
	}

	return &source{
		wg:               &sync.WaitGroup{},
		stop:             &atomic.Bool{},
		mutex:            &sync.RWMutex{},
		connection:       connection,
		subscribtions:    Sources.CreateSubscriptions(),
		subscrubtionStop: map[string]*atomic.Bool{}}, nil
}

func (ibk *source) GetContract(stockId string) *Market.Contract {
	contract, _, e := ibk.qualify(stockId)

	if e != nil {
		Logger.Error(Sources.SubscriptionError, stockId, e)
	}

	return contract
}

func (ibk *source) qualify(stockId string) (*Market.Contract, ibkr.Contract, error) {
	details, found, e := ibk.connection.GetContract(stockId)

	if e != nil {
		return nil, details.Contract, e
	}

	if !found {
		return nil, details.Contract, errors.New(ContractNotFoundError.Render(false, stockId))
	}

	tradingHours, tradingError := parseSchedule(stockId, "trading", details.TradingHours, details.TimeZoneID)
	liquidHours, liquidError := parseSchedule(stockId, "liquid", details.LiquidHours, details.TimeZoneID)

	contract := &Market.Contract{
		Stock: Market.Stock{
			Id:      stockId,
			Type:    details.StockType,
			MinTick: details.MinTick,
		},
		Id:              details.MarketName,
		Name:            details.LongName,
		Exchange:        details.Exchange,
		PrimaryExchange: details.PrimaryExchange,
		Country:         Market.CountryOf(details.PrimaryExchange, isin(details.SecurityIDs)),
		Currency:        details.Currency,
		Industry:        details.Industry,
		Category:        details.Category,
		Subcategory:     details.Subcategory,
		TradingHours:    tradingHours,
		LiquidHours:     liquidHours,
	}

	return contract, details.Contract, errors.Join(tradingError, liquidError)
}

func isin(ids []ibkr.TagValue) string {
	for _, id := range ids {
		if strings.EqualFold(id.Tag, "ISIN") {
			return id.Value
		}
	}

	return ""
}

func parseSchedule(stockId, kind, raw, timeZoneId string) (Schedule.Schedule, error) {
	schedule, e := Schedule.Parse(raw, timeZoneId)

	if e != nil {
		return Schedule.Schedule{}, errors.New(ScheduleParseError.Render(false, kind, stockId, e))
	}

	return *schedule, nil
}

func (ibk *source) metric(contract *Market.Contract, raw ibkr.Contract) (Market.TradeMetric, bool, error) {
	metric := Market.TradeMetric{Contract: contract}
	trades, e := ibk.connection.LastTicks(raw, ibkr.ShowTrades)

	if e != nil {
		return metric, false, errors.New(TicksError.Render(false, ibkr.ShowTrades, contract.Stock.Id, e))
	}

	quotes, e := ibk.connection.LastTicks(raw, ibkr.ShowBidAsk)

	if e != nil {
		return metric, false, errors.New(TicksError.Render(false, ibkr.ShowBidAsk, contract.Stock.Id, e))
	}

	if len(trades.Last) > 0 {
		trade := trades.Last[len(trades.Last)-1]

		metric.Time = trade.Time
		metric.Price = trade.Price
		metric.Size = trade.Size
	}

	if len(quotes.BidAsk) > 0 {
		quote := quotes.BidAsk[len(quotes.BidAsk)-1]

		metric.Bid = quote.BidPrice
		metric.Ask = quote.AskPrice
		metric.BidSize = quote.BidSize
		metric.AskSize = quote.AskSize

		if quote.Time.After(metric.Time) {
			metric.Time = quote.Time
		}
	}

	return metric, !metric.Time.IsZero(), nil
}

func (ibk *source) Subscribe(contract *Market.Contract, onPost Sources.OnPostHanlder, onError Sources.OnErrorHandler, updateInterval int) {
	ibk.mutex.Lock()
	defer ibk.mutex.Unlock()

	subscribtion := ibk.subscribtions.Subscribe(contract, onPost, onError)
	unsubscribed := &atomic.Bool{}
	ibk.subscrubtionStop[contract.Stock.Id] = unsubscribed

	ibk.wg.Add(1)

	go func(subscribtion *Sources.Subscription, wg *sync.WaitGroup, stop *atomic.Bool, unsubscribed *atomic.Bool) {
		defer wg.Done()

		interval := time.Duration(updateInterval) * time.Second
		current := &subscribtion.Contract
		raw := ibkr.Contract{}
		qualified := false

		for !stop.Load() && !unsubscribed.Load() {
			if !qualified || !current.IsActual() {
				fresh, freshRaw, e := ibk.qualify(subscribtion.Stock.Id)

				if fresh != nil {
					current, raw, qualified = fresh, freshRaw, true
				}

				if e != nil {
					subscribtion.Error(e)
				}
			}

			if qualified {
				if current.IsTradingHour() {
					metric, ok, e := ibk.metric(current, raw)

					if e != nil {
						subscribtion.Error(e)
					} else if ok {
						subscribtion.Post(metric)
					}
				} else {
					Logger.Debug("Is not working hour")
				}
			}

			pause(interval, stop, unsubscribed)
		}

		Logger.Info("Stop watching for the %s", subscribtion.Stock.Id)
	}(subscribtion, ibk.wg, ibk.stop, unsubscribed)
}

func (ibk *source) UnSubscribe(contract *Market.Contract) {
	ibk.subscribtions.UnSubscribe(contract)

	ibk.mutex.Lock()
	defer ibk.mutex.Unlock()

	ibk.subscrubtionStop[contract.Stock.Id].Store(true)

	delete(ibk.subscrubtionStop, contract.Stock.Id)
}

func (ibk *source) Close() {
	ibk.stop.Store(true)
	ibk.wg.Wait()
	ibk.subscribtions = Sources.CreateSubscriptions()
	ibk.connection.Close()
}

func (ibk *source) Run() {
	ibk.wg.Wait()
}

func pause(interval time.Duration, flags ...*atomic.Bool) {
	deadline := time.Now().Add(interval)

	for time.Now().Before(deadline) {
		for _, flag := range flags {
			if flag.Load() {
				return
			}
		}

		time.Sleep(min(PAUSE_STEP, time.Until(deadline)))
	}
}
