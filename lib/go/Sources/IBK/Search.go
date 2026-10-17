package IBK

import (
	"errors"

	"github.com/ThomasMarcelis/ibkr-go/v2"
	"github.com/grdvsng/stock-analysis/Data/Market"
)

func (ibk *source) Search(pattern string) ([]Market.Match, error) {
	symbols, e := ibk.connection.Search(pattern)

	if e != nil {
		return nil, errors.New(IBKApiError.Render(false, e))
	}

	matches := make([]Market.Match, 0, len(symbols))

	for _, symbol := range symbols {
		if symbol.SecType != ibkr.SecTypeStock {
			continue
		}

		matches = append(matches, Market.Match{
			Symbol:          symbol.Symbol,
			Name:            symbol.Description,
			PrimaryExchange: symbol.PrimaryExchange,
			Currency:        symbol.Currency,
			Country:         Market.CountryOf(symbol.PrimaryExchange, ""),
		})
	}

	return matches, nil
}
