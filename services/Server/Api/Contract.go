package Api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Control"
	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

type contractResponse struct {
	Contract *Market.Contract `json:"contract"`
	Metric   *Config.Metric   `json:"metric"`
}

func (api *Api) getContract(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	metric, e := Config.LoadMetric(api.db, id)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	contract, e := api.storedContract(id)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	if contract == nil {
		source, e := api.lookupSource()

		if e != nil {
			fail(w, r, http.StatusBadGateway, errors.New(SourceError.Render(false, e)))

			return
		}

		contract = source.GetContract(id)
	}

	if contract == nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("contract %s %w", id, ErrNotFound))

		return
	}

	reply(w, http.StatusOK, contractResponse{Contract: contract, Metric: metric})
}

func (api *Api) postContract(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	metric := Config.Metric{}

	if e := json.NewDecoder(r.Body).Decode(&metric); e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return
	}

	metric.ID = id

	if e := metric.Validate(); e != nil {
		fail(w, r, http.StatusBadRequest, e)

		return
	}

	if e := Config.SaveMetric(api.db, metric); e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	saved, e := Config.LoadMetric(api.db, id)

	if e != nil || saved == nil {
		saved = &metric
	}

	api.signal(Control.METRIC_UPDATED, id)

	reply(w, http.StatusCreated, saved)
}

func (api *Api) deleteContract(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredId(w, r)

	if !ok {
		return
	}

	metric, e := Config.LoadMetric(api.db, id)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	if metric == nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("subscription %s %w", id, ErrNotFound))

		return
	}

	if e := Config.RemoveMetric(api.db, id); e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	api.signal(Control.METRIC_REMOVED, id)

	w.WriteHeader(http.StatusNoContent)
}

func (api *Api) storedContract(id string) (*Market.Contract, error) {
	contracts, e := latest[Tables.Contract](api.db, Tables.CONTRACTS_TABLE, "stock_id", id)

	if e != nil || len(contracts) == 0 {
		return nil, e
	}

	stocks, e := latest[Tables.Stock](api.db, Tables.STOCKS_TABLE, "id", id)

	if e != nil {
		return nil, e
	}

	stock := Tables.Stock{Id: id}

	if len(stocks) > 0 {
		stock = stocks[0]
	}

	return contracts[0].Market(stock), nil
}

func latest[T Storage.Model](db *Storage.QuestDB, table, key, id string) ([]T, error) {
	where := ""

	if id != "" {
		where = fmt.Sprintf(" WHERE %s = %s", key, Storage.Text(id))
	}

	exists, e := db.Exists(*new(T))

	if e != nil || !exists {
		return nil, e
	}

	return Storage.Select[T](db, fmt.Sprintf("SELECT * FROM %s%s LATEST ON updated_at PARTITION BY %s", table, where, key))
}
