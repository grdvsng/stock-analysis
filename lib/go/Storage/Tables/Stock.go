package Tables

import (
	"time"

	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/shopspring/decimal"
)

const STOCKS_TABLE = "stocks"

type Stock struct {
	Id        string          `json:"id" db:"id,dedup" dbtype:"SYMBOL"`
	Type      string          `json:"type" db:"type" dbtype:"SYMBOL"`
	MinTick   decimal.Decimal `json:"min_tick" db:"min_tick"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at,designated"`
}

func (Stock) Table() string {
	return STOCKS_TABLE
}

func NewStock(stock Market.Stock, updatedAt time.Time) Stock {
	return Stock{
		Id:        stock.Id,
		Type:      stock.Type,
		MinTick:   stock.MinTick,
		UpdatedAt: updatedAt,
	}
}

func (stock Stock) Market() Market.Stock {
	return Market.Stock{
		Id:      stock.Id,
		Type:    stock.Type,
		MinTick: stock.MinTick,
	}
}
