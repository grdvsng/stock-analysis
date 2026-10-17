package Collector

import (
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

func (collector *Collector) Cleanup() {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	now := time.Now()
	tables, e := collector.db.Tables(Tables.METRICS_TABLE_PREFIX)

	if e != nil {
		Logger.Error(StorageError, Tables.METRICS_TABLE_PREFIX, e)

		return
	}

	active := make(map[string]Config.Metric, len(collector.stocks))

	for id, stock := range collector.stocks {
		active[Tables.MetricsTable(id)] = stock
	}

	for _, table := range tables {
		if stock, exists := active[table]; exists {
			if e := collector.db.DropPartitionsBefore(Tables.Metric{StockId: stock.ID}, now.Add(-stock.Retention())); e != nil {
				Logger.Error(StorageError, stock.ID, e)
			}
		} else {
			if e := collector.db.Drop(table); e != nil {
				Logger.Error(StorageError, table, e)
			} else {
				Logger.Info("Dropped %s", table)
			}
		}
	}

	for id, current := range collector.subscriptions {
		if e := collector.saveContract(current.contract, now); e != nil {
			Logger.Error(StorageError, id, e)

			return
		}

		current.updatedAt = now
	}

	for _, model := range []Storage.Model{Tables.Stock{}, Tables.Contract{}} {
		if e := collector.db.DropPartitionsBefore(model, now); e != nil {
			Logger.Error(StorageError, model.Table(), e)
		}
	}

	Logger.Info("Cleanup finished")
}
