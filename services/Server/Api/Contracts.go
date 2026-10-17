package Api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

const DATE_FORMAT = "2006-01-02"

type historyResponse struct {
	From    time.Time                  `json:"from"`
	To      time.Time                  `json:"to"`
	Metrics map[string][]Tables.Metric `json:"metrics"`
}

func (api *Api) getContracts(w http.ResponseWriter, r *http.Request) {
	metrics, _, e := Config.LoadMetrics(api.db)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	contracts, e := latest[Tables.Contract](api.db, Tables.CONTRACTS_TABLE, "stock_id", "")

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	stocks, e := latest[Tables.Stock](api.db, Tables.STOCKS_TABLE, "id", "")

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	byContract := map[string]Tables.Contract{}
	byStock := map[string]Tables.Stock{}

	for _, contract := range contracts {
		byContract[contract.StockId] = contract
	}

	for _, stock := range stocks {
		byStock[stock.Id] = stock
	}

	result := make([]contractResponse, 0, len(metrics))

	for i := range metrics {
		item := contractResponse{Metric: &metrics[i]}

		if contract, exists := byContract[metrics[i].ID]; exists {
			stock, found := byStock[metrics[i].ID]

			if !found {
				stock = Tables.Stock{Id: metrics[i].ID}
			}

			item.Contract = contract.Market(stock)
		}

		result = append(result, item)
	}

	reply(w, http.StatusOK, result)
}

func (api *Api) searchContracts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	if query == "" {
		fail(w, r, http.StatusBadRequest, errors.New("query required"))

		return
	}

	source, e := api.lookupSource()

	if e != nil {
		fail(w, r, http.StatusBadGateway, errors.New(SourceError.Render(false, e)))

		return
	}

	matches, e := source.Search(query)

	if e != nil {
		fail(w, r, http.StatusBadGateway, e)

		return
	}

	reply(w, http.StatusOK, matches)
}

func (api *Api) getHistory(w http.ResponseWriter, r *http.Request) {
	from, e := parseTime(r.URL.Query().Get("from"))

	if e != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("from: %w", e))

		return
	}

	to := time.Now().UTC()

	if raw := r.URL.Query().Get("to"); raw != "" {
		if to, e = parseTime(raw); e != nil {
			fail(w, r, http.StatusBadRequest, fmt.Errorf("to: %w", e))

			return
		}
	}

	metrics, e := metricsBetween(api.db, from, to)

	if e != nil {
		fail(w, r, http.StatusInternalServerError, e)

		return
	}

	reply(w, http.StatusOK, historyResponse{From: from, To: to, Metrics: metrics})
}

func metricsBetween(db *Storage.QuestDB, from, to time.Time) (map[string][]Tables.Metric, error) {
	tables, e := db.Tables(Tables.METRICS_TABLE_PREFIX)

	if e != nil {
		return nil, e
	}

	result := map[string][]Tables.Metric{}

	for _, table := range tables {
		rows, e := Storage.Select[Tables.Metric](db, fmt.Sprintf(
			"SELECT * FROM %s WHERE time >= %s AND time <= %s ORDER BY time",
			table,
			Storage.Timestamp(from),
			Storage.Timestamp(to),
		))

		if e != nil {
			return nil, e
		}

		for _, row := range rows {
			result[row.StockId] = append(result[row.StockId], row)
		}
	}

	return result, nil
}

func parseTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("required")
	}

	if moment, e := time.Parse(time.RFC3339Nano, raw); e == nil {
		return moment.UTC(), nil
	}

	moment, e := time.Parse(DATE_FORMAT, raw)

	return moment.UTC(), e
}
