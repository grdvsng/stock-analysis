package Api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Indicators"
	"github.com/grdvsng/stock-analysis/Server/Monitors"
	"github.com/grdvsng/stock-analysis/Time/Duration"
)

type monitorRequest struct {
	Name          string             `json:"name"`
	Stocks        []string           `json:"stocks"`
	Notifications []string           `json:"notifications"`
	Interval      string             `json:"interval"`
	EventLifetime *Duration.Duration `json:"event_lifetime"`
	Alerts        []Monitors.Alert   `json:"alerts"`
}

func (api *Api) getSettings(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, map[string]any{"ws_port": api.wsPort})
}

func (api *Api) getIndicators(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, Indicators.Catalog())
}

func (api *Api) getMonitors(w http.ResponseWriter, r *http.Request) {
	monitors, e := Monitors.List(api.db)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	reply(w, http.StatusOK, monitors)
}

func (api *Api) postMonitor(w http.ResponseWriter, r *http.Request) {
	monitor, ok := api.decodeMonitor(w, r, Monitors.Monitor{Id: Monitors.NewId()})

	if ok {
		api.saveMonitor(w, r, monitor, http.StatusCreated)
	}
}

func (api *Api) putMonitor(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	current, e := Monitors.Get(api.db, id)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	if current == nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("monitor %s %w", id, ErrNotFound))

		return
	}

	if monitor, ok := api.decodeMonitor(w, r, *current); ok {
		api.saveMonitor(w, r, monitor, http.StatusOK)
	}
}

func (api *Api) deleteMonitor(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	current, e := Monitors.Get(api.db, id)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	if current == nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("monitor %s %w", id, ErrNotFound))

		return
	}

	if e := Monitors.Remove(api.db, id); e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *Api) decodeMonitor(w http.ResponseWriter, r *http.Request, monitor Monitors.Monitor) (Monitors.Monitor, bool) {
	body := monitorRequest{Name: monitor.Name, Stocks: monitor.Stocks, Notifications: monitor.Notifications, Interval: monitor.Interval, EventLifetime: monitor.EventLifetime, Alerts: monitor.Alerts}

	if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return monitor, false
	}

	if body.Interval == "" {
		body.Interval = Monitors.DEFAULT_INTERVAL
	}

	monitor.Name, monitor.Stocks, monitor.Notifications, monitor.Interval = body.Name, body.Stocks, body.Notifications, body.Interval
	monitor.EventLifetime = body.EventLifetime
	monitor.Alerts = body.Alerts
	monitor.SanitizeAlerts()

	if e := monitor.Validate(); e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return monitor, false
	}

	metrics, _, e := Config.LoadMetrics(api.db)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return monitor, false
	}

	subscribed := map[string]bool{}

	for _, metric := range metrics {
		subscribed[metric.ID] = true
	}

	for _, stock := range monitor.Stocks {
		if !subscribed[stock] {
			fail(w, r, http.StatusBadRequest, fmt.Errorf("%s is not among subscribed contracts", stock))

			return monitor, false
		}
	}

	return monitor, true
}

func (api *Api) saveMonitor(w http.ResponseWriter, r *http.Request, monitor Monitors.Monitor, status int) {
	saved, e := Monitors.Save(api.db, monitor)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	reply(w, status, saved)
}
