package Config

import "github.com/grdvsng/stock-analysis/Logger"

type logging struct {
	Level Logger.Level `json:"level"`
	Path  string       `json:"path"`
}
