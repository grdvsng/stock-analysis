package Api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Control"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Storage"
)

const SIGNAL_TIMEOUT = 2 * time.Minute

var ErrNotFound = errors.New("not found")

type Api struct {
	db              *Storage.QuestDB
	collectorConfig string
	collector       *Config.Config
	control         *Control.Client
	updateTimeout   time.Duration
	uiPath          string
	wsPort          int
	mutex           *sync.Mutex
	source          Sources.Source
}

func New(db *Storage.QuestDB, collectorConfig string, collector *Config.Config, control *Control.Client, updateTimeout time.Duration, uiPath string, wsPort int) *Api {
	return &Api{
		db:              db,
		collectorConfig: collectorConfig,
		collector:       collector,
		control:         control,
		updateTimeout:   updateTimeout,
		uiPath:          uiPath,
		wsPort:          wsPort,
		mutex:           &sync.Mutex{},
	}
}

func (api *Api) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /config", api.getConfig)
	mux.HandleFunc("POST /config", api.postConfig)
	mux.HandleFunc("GET /contract", api.getContract)
	mux.HandleFunc("POST /contract", api.postContract)
	mux.HandleFunc("DELETE /contract", api.deleteContract)
	mux.HandleFunc("GET /contracts", api.getContracts)
	mux.HandleFunc("GET /contracts/history", api.getHistory)
	mux.HandleFunc("GET /contracts/search", api.searchContracts)
	mux.HandleFunc("GET /settings", api.getSettings)
	mux.HandleFunc("GET /indicators", api.getIndicators)
	mux.HandleFunc("GET /forecast", api.getForecast)
	mux.HandleFunc("GET /monitors", api.getMonitors)
	mux.HandleFunc("POST /monitors", api.postMonitor)
	mux.HandleFunc("PUT /monitor", api.putMonitor)
	mux.HandleFunc("DELETE /monitor", api.deleteMonitor)

	if api.uiPath != "" {
		mux.Handle("GET /", ui(api.uiPath))
	}

	return mux
}

func (api *Api) Close() {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.source != nil {
		api.source.Close()
		api.source = nil
	}
}

func (api *Api) reloadCollector() {
	collector, e := Config.Load(api.collectorConfig)

	if e != nil {
		return
	}

	api.mutex.Lock()
	defer api.mutex.Unlock()

	api.collector = collector

	if api.source != nil {
		api.source.Close()
		api.source = nil
	}
}

func (api *Api) signal(kind Control.Type, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), SIGNAL_TIMEOUT)
	defer cancel()

	reply, e := api.control.Request(ctx, Control.Message{Type: kind, Metric: id})

	if e == nil && !reply.Ok {
		e = errors.New(reply.Error)
	}

	if e != nil {
		Logger.Warning(SignalWarning, kind, id, e)
	}
}

func (api *Api) lookupSource() (Sources.Source, error) {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.source == nil {
		source, e := api.collector.Source.Gateway(api.collector.Workers)

		if e != nil {
			return nil, e
		}

		api.source = source
	}

	return api.source, nil
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}

func fail(w http.ResponseWriter, r *http.Request, status int, e error) {
	if status >= http.StatusInternalServerError {
		Logger.Error(RequestError, r.Method, r.URL.Path, e)
	}

	reply(w, status, map[string]string{"error": e.Error()})
}

func requiredId(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.URL.Query().Get("id")

	if id == "" {
		fail(w, r, http.StatusBadRequest, errors.New("id required"))
	}

	return id, id != ""
}
