package Monitors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grdvsng/stock-analysis/Indicators"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

const (
	DEFAULT_INTERVAL = "1m"
	MAX_WARMUP       = 1000
)

var Intervals = map[string]time.Duration{
	"1m":  time.Minute,
	"5m":  5 * time.Minute,
	"15m": 15 * time.Minute,
	"30m": 30 * time.Minute,
	"1h":  time.Hour,
	"1d":  24 * time.Hour,
}

const candleColumns = "time, first(price::double) open, max(price::double) high, min(price::double) low, last(price::double) close, sum(size::double) volume"

func Candles(db *Storage.QuestDB, stockId, interval string, from, to time.Time, warmup int) ([]Indicators.Candle, error) {
	table := Tables.MetricsTable(stockId)
	exists, e := db.Exists(Tables.Metric{StockId: stockId})

	if e != nil || !exists {
		return nil, e
	}

	before := []Indicators.Candle{}

	if warmup > 0 {
		before, e = query(db, fmt.Sprintf(
			"SELECT %s FROM %s WHERE price != null AND time < %s SAMPLE BY %s ALIGN TO CALENDAR LIMIT -%d",
			candleColumns, table, Storage.Timestamp(from), interval, min(warmup, MAX_WARMUP),
		))

		if e != nil {
			return nil, e
		}
	}

	upper := ""

	if !to.IsZero() {
		upper = " AND time <= " + Storage.Timestamp(to)
	}

	inside, e := query(db, fmt.Sprintf(
		"SELECT %s FROM %s WHERE price != null AND time >= %s%s SAMPLE BY %s ALIGN TO CALENDAR",
		candleColumns, table, Storage.Timestamp(from), upper, interval,
	))

	if e != nil {
		return nil, e
	}

	return append(before, inside...), nil
}

func query(db *Storage.QuestDB, sql string) ([]Indicators.Candle, error) {
	result, e := db.Exec(sql)

	if e != nil {
		return nil, e
	}

	candles := make([]Indicators.Candle, 0, len(result.Dataset))

	for _, row := range result.Dataset {
		if len(row) < 6 {
			continue
		}

		var moment time.Time
		var values [5]*float64

		if json.Unmarshal(row[0], &moment) != nil {
			continue
		}

		for i := range values {
			json.Unmarshal(row[i+1], &values[i])
		}

		if values[0] == nil || values[1] == nil || values[2] == nil || values[3] == nil {
			continue
		}

		candle := Indicators.Candle{Time: moment, Open: *values[0], High: *values[1], Low: *values[2], Close: *values[3]}

		if values[4] != nil {
			candle.Volume = *values[4]
		}

		candles = append(candles, candle)
	}

	return candles, nil
}
