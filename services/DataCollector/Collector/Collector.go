package Collector

import (
	"context"
	"sync"
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

type subscription struct {
	stock     Config.Metric
	source    Sources.Source
	contract  *Market.Contract
	updatedAt time.Time
}

type Collector struct {
	ctx           context.Context
	cancel        context.CancelFunc
	reprocessing  *sync.WaitGroup
	mutex         *sync.Mutex
	db            *Storage.QuestDB
	sources       []Sources.Source
	next          int
	stocks        map[string]Config.Metric
	subscriptions map[string]*subscription
}

func New(db *Storage.QuestDB, sources []Sources.Source) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		ctx:           ctx,
		cancel:        cancel,
		reprocessing:  &sync.WaitGroup{},
		mutex:         &sync.Mutex{},
		db:            db,
		sources:       sources,
		stocks:        map[string]Config.Metric{},
		subscriptions: map[string]*subscription{},
	}
}

func (collector *Collector) Sync(metrics []Config.Metric) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	now := time.Now()
	wanted := make(map[string]Config.Metric, len(metrics))

	for _, metric := range metrics {
		wanted[metric.ID] = metric
	}

	for id, current := range collector.subscriptions {
		stock, exists := wanted[id]

		if !exists || stock.Interval() != current.stock.Interval() {
			collector.unsubscribe(id)
		} else {
			current.stock = stock
		}
	}

	previous := collector.stocks
	collector.stocks = wanted

	for id, metric := range wanted {
		if _, subscribed := collector.subscriptions[id]; subscribed {
			continue
		}

		if known, exists := previous[id]; exists && known.Interval() == metric.Interval() {
			continue
		}

		collector.subscribe(metric, now)
	}
}

func (collector *Collector) Resubscribe(id string, metrics []Config.Metric) {
	collector.mutex.Lock()

	if _, subscribed := collector.subscriptions[id]; subscribed {
		collector.unsubscribe(id)
	}

	delete(collector.stocks, id)
	collector.mutex.Unlock()

	collector.Sync(metrics)
}

func (collector *Collector) Remove(id string, metrics []Config.Metric) error {
	collector.Resubscribe(id, metrics)

	for _, metric := range metrics {
		if metric.ID == id {
			return nil
		}
	}

	return collector.db.Drop(Tables.MetricsTable(id))
}

func (collector *Collector) Refresh() {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	collector.refresh()
}

func (collector *Collector) Close() {
	collector.cancel()
	collector.reprocessing.Wait()

	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	for id := range collector.subscriptions {
		collector.unsubscribe(id)
	}
}

func (collector *Collector) refresh() {
	now := time.Now()

	for id, stock := range collector.stocks {
		current, subscribed := collector.subscriptions[id]

		if !subscribed {
			collector.subscribe(stock, now)
		} else if now.Sub(current.updatedAt) >= stock.ContractInterval() || !current.contract.IsActual() {
			collector.update(current, now)
		}
	}
}

func (collector *Collector) subscribe(stock Config.Metric, now time.Time) {
	source := collector.sources[collector.next%len(collector.sources)]
	contract := source.GetContract(stock.ID)

	if contract == nil {
		Logger.Error(ContractNotFoundError, stock.ID)

		return
	}

	if e := collector.saveContract(contract, now); e != nil {
		Logger.Error(StorageError, stock.ID, e)

		return
	}

	fresh, e := collector.isNew(stock.ID)

	if e != nil {
		Logger.Error(StorageError, stock.ID, e)

		return
	}

	if e := collector.db.Create(Tables.Metric{StockId: stock.ID}); e != nil {
		Logger.Error(StorageError, stock.ID, e)

		return
	}

	collector.next++
	collector.subscriptions[stock.ID] = &subscription{
		stock:     stock,
		source:    source,
		contract:  contract,
		updatedAt: now,
	}

	source.Subscribe(contract, collector.onPost, collector.onError(stock.ID), seconds(stock.Interval()))
	Logger.Info("Subscribed to %s", stock.ID)

	if fresh && stock.Reprocess > 0 {
		collector.reprocessing.Add(1)

		go collector.reprocess(source, contract, stock.Reprocess)
	}
}

func (collector *Collector) isNew(stockId string) (bool, error) {
	metric := Tables.Metric{StockId: stockId}
	exists, e := collector.db.Exists(metric)

	if e != nil || !exists {
		return !exists, e
	}

	count, e := collector.db.Count(metric)

	return count == 0, e
}

func (collector *Collector) reprocess(source Sources.Source, contract *Market.Contract, days int) {
	defer collector.reprocessing.Done()

	Logger.Info("Reprocess %s for %d days", contract.Stock.Id, days)

	saved := 0
	collectedAt := time.Now()
	e := source.History(collector.ctx, contract, days, func(metrics []Market.TradeMetric) {
		records := make([]Storage.Model, 0, len(metrics))

		for _, metric := range metrics {
			records = append(records, Tables.NewMetric(metric, collectedAt))
		}

		if e := collector.db.Insert(records...); e != nil {
			Logger.Error(StorageError, contract.Stock.Id, e)
		} else {
			saved += len(records)
		}
	})

	if e != nil {
		Logger.Error(ReprocessError, contract.Stock.Id, e)
	}

	Logger.Info("Reprocess %s finished, %d metrics saved", contract.Stock.Id, saved)
}

func (collector *Collector) update(current *subscription, now time.Time) {
	contract := current.source.GetContract(current.stock.ID)

	if contract == nil {
		Logger.Error(ContractNotFoundError, current.stock.ID)

		return
	}

	if e := collector.saveContract(contract, now); e != nil {
		Logger.Error(StorageError, current.stock.ID, e)

		return
	}

	current.contract = contract
	current.updatedAt = now
}

func (collector *Collector) unsubscribe(id string) {
	current := collector.subscriptions[id]

	current.source.UnSubscribe(current.contract)
	delete(collector.subscriptions, id)
	Logger.Info("Unsubscribed from %s", id)
}

func (collector *Collector) onPost(metric Market.TradeMetric) {
	if e := collector.db.Insert(Tables.NewMetric(metric, time.Now())); e != nil {
		Logger.Error(StorageError, metric.Contract.Stock.Id, e)
	}
}

func (collector *Collector) onError(stockId string) Sources.OnErrorHandler {
	return func(e error) {
		Logger.Error(SourceError, stockId, e)
	}
}

func (collector *Collector) saveContract(contract *Market.Contract, now time.Time) error {
	return collector.db.Insert(Tables.NewStock(contract.Stock, now), Tables.NewContract(contract, now))
}

func seconds(duration time.Duration) int {
	return max(1, int(duration.Seconds()))
}
