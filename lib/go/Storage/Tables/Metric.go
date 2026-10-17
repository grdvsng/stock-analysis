package Tables

import (
	"regexp"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/shopspring/decimal"
)

const METRICS_TABLE_PREFIX = "metrics_"

var notTableChar = regexp.MustCompile(`[^a-z0-9]+`)

type Metric struct {
	StockId     string           `json:"stock_id" db:"stock_id" dbtype:"SYMBOL"`
	ContractId  string           `json:"contract_id" db:"contract_id" dbtype:"SYMBOL"`
	Time        time.Time        `json:"time" db:"time,designated,dedup"`
	Price       *decimal.Decimal `json:"price" db:"price"`
	Size        *decimal.Decimal `json:"size" db:"size"`
	Bid         *decimal.Decimal `json:"bid" db:"bid"`
	Ask         *decimal.Decimal `json:"ask" db:"ask"`
	BidSize     *decimal.Decimal `json:"bid_size" db:"bid_size"`
	AskSize     *decimal.Decimal `json:"ask_size" db:"ask_size"`
	CollectedAt time.Time        `json:"collected_at" db:"collected_at"`
}

func MetricsTable(stockId string) string {
	return METRICS_TABLE_PREFIX + notTableChar.ReplaceAllString(strings.ToLower(stockId), "_")
}

func (metric Metric) Table() string {
	return MetricsTable(metric.StockId)
}

func NewMetric(metric Market.TradeMetric, collectedAt time.Time) Metric {
	return Metric{
		StockId:     metric.Contract.Stock.Id,
		ContractId:  metric.Contract.Id,
		Time:        metric.Time,
		Price:       optional(metric.Price),
		Size:        optional(metric.Size),
		Bid:         optional(metric.Bid),
		Ask:         optional(metric.Ask),
		BidSize:     optional(metric.BidSize),
		AskSize:     optional(metric.AskSize),
		CollectedAt: collectedAt,
	}
}

func optional(value decimal.Decimal) *decimal.Decimal {
	if value.IsZero() {
		return nil
	}

	return &value
}
