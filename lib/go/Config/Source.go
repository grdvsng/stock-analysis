package Config

import (
	"fmt"

	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Sources/IBK"
)

type sourceType string

const (
	IBKSource sourceType = "ibk"
)

type source struct {
	Host     string     `json:"host"`
	Port     int        `json:"port"`
	Timeout  int        `json:"timeout"`
	ClientId int        `json:"client_id"`
	Type     sourceType `json:"type"`
}

func (s *source) Gateway(worker int) (source Sources.Source, err error) {
	switch s.Type {
	case IBKSource:
		source, err = IBK.Connect(s.Host, s.Port, s.Timeout, s.ClientId+worker)
	default:
		err = fmt.Errorf("Unknown source %s", s.Type)
	}

	return source, err
}
