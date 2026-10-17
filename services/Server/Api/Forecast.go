package Api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/grdvsng/stock-analysis/Indicators"
	"github.com/grdvsng/stock-analysis/Server/Monitors"
)

const (
	FORECAST_CANDLES = 300
	FORECAST_CHART   = 150
)

type forecastResponse struct {
	Stock    string              `json:"stock"`
	Name     string              `json:"name"`
	Interval string              `json:"interval"`
	Trading  bool                `json:"trading"`
	Forecast Indicators.Forecast `json:"forecast"`
	Candles  []Indicators.Candle `json:"candles"`
}

func (api *Api) getForecast(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	interval := r.URL.Query().Get("interval")

	if interval == "" {
		interval = "5m"
	}

	if _, exists := Monitors.Intervals[interval]; !exists {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("unknown interval %q", interval))

		return
	}

	candles, e := Monitors.Candles(api.db, id, interval, time.Now().UTC().Add(Monitors.Intervals[interval]), time.Time{}, FORECAST_CANDLES)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	forecast, e := Indicators.Predict(candles)

	if errors.Is(e, Indicators.ErrNotEnoughCandles) {
		fail(w, r, http.StatusUnprocessableEntity, fmt.Errorf("%s has %d %s candles, at least %d are needed", id, len(candles), interval, Indicators.FORECAST_MIN_CANDLES))

		return
	}

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	response := forecastResponse{Stock: id, Interval: interval, Forecast: forecast, Candles: candles[max(0, len(candles)-FORECAST_CHART):]}

	if contract, e := api.storedContract(id); e == nil && contract != nil {
		response.Name = contract.Name
		response.Trading = contract.IsTradingHour()
	}

	reply(w, http.StatusOK, response)
}
