package Market

import "github.com/grdvsng/stock-analysis/Time/Schedule"

type Contract struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Exchange    string `json:"exchange"`
	Currency    string `json:"currency"`
	Industry    string `json:"industry"`
	Category    string `json:"category"`
	Subcategory string `json:"sub_category"`

	PrimaryExchange string `json:"primary_exchange"`
	Country         string `json:"country"`

	TradingHours Schedule.Schedule `json:"trading_hours"`
	LiquidHours  Schedule.Schedule `json:"liquid_hours"`

	Stock Stock `json:"stock"`
}

func (contract *Contract) IsTradingHour() bool {
	return contract.TradingHours.Available()
}

func (contract *Contract) IsLiquidHours() bool {
	return contract.LiquidHours.Available()
}

func (contract *Contract) IsActual() bool {
	return contract.TradingHours.Actual()
}
