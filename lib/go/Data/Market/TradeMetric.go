package Market

import (
	"time"

	"github.com/shopspring/decimal"
)

type TradeMetric struct {
	Contract *Contract
	Time     time.Time
	Price    decimal.Decimal
	Size     decimal.Decimal
	Bid      decimal.Decimal
	Ask      decimal.Decimal
	BidSize  decimal.Decimal
	AskSize  decimal.Decimal
}
