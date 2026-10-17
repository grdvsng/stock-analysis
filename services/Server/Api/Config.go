package Api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Control"
)

const MAX_CONFIG_SIZE = 1 << 20

func (api *Api) getConfig(w http.ResponseWriter, r *http.Request) {
	raw, e := os.ReadFile(api.collectorConfig)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (api *Api) postConfig(w http.ResponseWriter, r *http.Request) {
	raw, e := io.ReadAll(io.LimitReader(r.Body, MAX_CONFIG_SIZE))

	if e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return
	}

	if _, e := Config.Validate(raw, api.collectorConfig); e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), api.updateTimeout)
	defer cancel()

	reply, e := api.control.Request(ctx, Control.Message{Type: Control.UPDATE_CONFIG, Config: raw})

	if errors.Is(e, Control.ErrDisconnected) {
		fail(w, r, http.StatusServiceUnavailable, e)

		return
	}

	if e != nil {
		fail(w, r, http.StatusGatewayTimeout, e)

		return
	}

	if !reply.Ok {
		fail(w, r, http.StatusBadRequest, errors.New(reply.Error))

		return
	}

	api.reloadCollector()

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}
