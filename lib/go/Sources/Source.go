package Sources

import (
	"context"

	"github.com/grdvsng/stock-analysis/Data/Market"
)

type OnBatchHandler func(metrics []Market.TradeMetric)

type Source interface {
	GetContract(stockId string) *Market.Contract
	Subscribe(contract *Market.Contract, onPost OnPostHanlder, onError OnErrorHandler, updateInterval int)
	UnSubscribe(stock *Market.Contract)
	History(ctx context.Context, contract *Market.Contract, days int, onBatch OnBatchHandler) error
	Search(pattern string) ([]Market.Match, error)
	Close()
	Run()
}

type SourceFabric func(host string, port, timeout, clientId int) Source
