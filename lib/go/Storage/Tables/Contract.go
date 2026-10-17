package Tables

import (
	"time"

	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Time/Schedule"
)

const CONTRACTS_TABLE = "contracts"

type Contract struct {
	Id           string            `json:"id" db:"id" dbtype:"SYMBOL"`
	StockId      string            `json:"stock_id" db:"stock_id,dedup" dbtype:"SYMBOL"`
	Name         string            `json:"name" db:"name"`
	Exchange     string            `json:"exchange" db:"exchange" dbtype:"SYMBOL"`
	Currency     string            `json:"currency" db:"currency" dbtype:"SYMBOL"`
	Industry     string            `json:"industry" db:"industry" dbtype:"SYMBOL"`
	Category     string            `json:"category" db:"category" dbtype:"SYMBOL"`
	Subcategory  string            `json:"sub_category" db:"sub_category" dbtype:"SYMBOL"`
	Primary      string            `json:"primary_exchange" db:"primary_exchange" dbtype:"SYMBOL"`
	Country      string            `json:"country" db:"country" dbtype:"SYMBOL"`
	TimeZone     string            `json:"time_zone" db:"time_zone" dbtype:"SYMBOL"`
	TradingHours Schedule.Schedule `json:"trading_hours" db:"trading_hours"`
	LiquidHours  Schedule.Schedule `json:"liquid_hours" db:"liquid_hours"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at,designated"`
}

func (Contract) Table() string {
	return CONTRACTS_TABLE
}

func NewContract(contract *Market.Contract, updatedAt time.Time) Contract {
	return Contract{
		Id:           contract.Id,
		StockId:      contract.Stock.Id,
		Name:         contract.Name,
		Exchange:     contract.Exchange,
		Currency:     contract.Currency,
		Industry:     contract.Industry,
		Category:     contract.Category,
		Subcategory:  contract.Subcategory,
		Primary:      contract.PrimaryExchange,
		Country:      contract.Country,
		TimeZone:     timeZone(contract.TradingHours),
		TradingHours: contract.TradingHours,
		LiquidHours:  contract.LiquidHours,
		UpdatedAt:    updatedAt,
	}
}

func (contract Contract) Market(stock Stock) *Market.Contract {
	return &Market.Contract{
		Id:              contract.Id,
		Name:            contract.Name,
		Exchange:        contract.Exchange,
		Currency:        contract.Currency,
		Industry:        contract.Industry,
		Category:        contract.Category,
		Subcategory:     contract.Subcategory,
		PrimaryExchange: contract.Primary,
		Country:         contract.Country,
		TradingHours:    contract.TradingHours,
		LiquidHours:     contract.LiquidHours,
		Stock:           stock.Market(),
	}
}

func timeZone(schedule Schedule.Schedule) string {
	if schedule.Location == nil {
		return ""
	}

	return schedule.Location.Name
}
