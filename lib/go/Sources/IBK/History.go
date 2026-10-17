package IBK

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/ThomasMarcelis/ibkr-go/v2"
	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Time/Schedule"
)

const HISTORY_EXTRA_REQUESTS = 7

func (ibk *source) History(ctx context.Context, contract *Market.Contract, days int, onBatch Sources.OnBatchHandler) error {
	_, raw, e := ibk.qualify(contract.Stock.Id)

	if e != nil {
		return e
	}

	collected := map[string]bool{}
	end := time.Now().UTC()

	for request := 0; len(collected) < days && request < days+HISTORY_EXTRA_REQUESTS; request++ {
		if e := ctx.Err(); e != nil {
			return e
		}

		trades, e := ibk.connection.Bars(ctx, raw, end, ibkr.ShowTrades)

		if e != nil {
			return errors.New(TicksError.Render(false, ibkr.ShowTrades, contract.Stock.Id, e))
		}

		quotes, e := ibk.connection.Bars(ctx, raw, end, ibkr.ShowBidAsk)

		if e != nil {
			return errors.New(TicksError.Render(false, ibkr.ShowBidAsk, contract.Stock.Id, e))
		}

		metrics := barMetrics(contract, trades, quotes)

		if len(metrics) == 0 {
			return nil
		}

		end = metrics[0].Time
		collected[tradingDay(contract.TradingHours, end)] = true

		onBatch(metrics)
	}

	return nil
}

func barMetrics(contract *Market.Contract, trades, quotes []ibkr.Bar) []Market.TradeMetric {
	byTime := map[int64]*Market.TradeMetric{}

	at := func(moment time.Time) *Market.TradeMetric {
		key := moment.UnixMicro()

		if _, exists := byTime[key]; !exists {
			byTime[key] = &Market.TradeMetric{Contract: contract, Time: moment}
		}

		return byTime[key]
	}

	for _, bar := range trades {
		metric := at(bar.Time)
		metric.Price = bar.Close
		metric.Size = bar.Volume
	}

	for _, bar := range quotes {
		metric := at(bar.Time)
		metric.Bid = bar.Open
		metric.Ask = bar.Close
	}

	metrics := make([]Market.TradeMetric, 0, len(byTime))

	for _, metric := range byTime {
		metrics = append(metrics, *metric)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Time.Before(metrics[j].Time)
	})

	return metrics
}

func tradingDay(schedule Schedule.Schedule, moment time.Time) string {
	if schedule.Location != nil {
		moment = moment.In(&schedule.Location.Location)
	}

	return moment.Format(Schedule.YMD_FORMAT)
}
