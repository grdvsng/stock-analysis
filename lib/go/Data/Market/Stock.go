package Market

import "github.com/shopspring/decimal"

type Stock struct {
	Id      string          `json:"id"`
	Type    string          `json:"type"`
	MinTick decimal.Decimal `json:"min_tick"`
}
