package Api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Indicators"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Server/Monitors"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
	"github.com/grdvsng/stock-analysis/Time/Duration"
)

const (
	MAX_MONITOR_STOCKS = 50
	CONTRACT_REFRESH   = time.Minute
)

type monitorParams struct {
	stocks      []string
	interval    string
	definitions []Indicators.Definition
	lifetime    time.Duration
	from        time.Time
	to          time.Time
	warmup      int
}

type stockUpdate struct {
	changed bool
	Since   time.Time           `json:"since"`
	Trading bool                `json:"trading"`
	Candles []Indicators.Candle `json:"candles"`
	Signals []Indicators.Signal `json:"signals"`
}

type monitorMessage struct {
	Type   string                  `json:"type"`
	Error  string                  `json:"error,omitempty"`
	Stocks map[string]*stockUpdate `json:"stocks,omitempty"`
}

type stockState struct {
	candles   []Indicators.Candle
	contract  *Market.Contract
	refreshed time.Time
}

func parseMonitor(values url.Values) (monitorParams, error) {
	params := monitorParams{interval: values.Get("interval")}

	if params.interval == "" {
		params.interval = Monitors.DEFAULT_INTERVAL
	}

	if _, exists := Monitors.Intervals[params.interval]; !exists {
		return params, fmt.Errorf("unknown interval %q", params.interval)
	}

	for _, stock := range strings.Split(values.Get("stocks"), ",") {
		if stock = strings.TrimSpace(stock); stock != "" {
			params.stocks = append(params.stocks, stock)
		}
	}

	if len(params.stocks) > MAX_MONITOR_STOCKS {
		return params, fmt.Errorf("at most %d stocks", MAX_MONITOR_STOCKS)
	}

	for _, key := range strings.Split(values.Get("notifications"), ",") {
		if key = strings.TrimSpace(key); key == "" {
			continue
		}

		definition, exists := Indicators.Lookup(key)

		if !exists {
			return params, fmt.Errorf("unknown notification %q", key)
		}

		params.definitions = append(params.definitions, definition)
		params.warmup = max(params.warmup, definition.Warmup)
	}

	var e error

	lifetime := (*Duration.Duration)(nil)

	if raw := values.Get("lifetime"); raw != "" {
		lifetime = &Duration.Duration{}

		if e = lifetime.UnmarshalText([]byte(raw)); e != nil || lifetime.Duration() <= 0 {
			return params, fmt.Errorf("lifetime: invalid %q", raw)
		}
	}

	params.lifetime = Monitors.Lifetime(lifetime, params.interval)

	if raw := values.Get("from"); raw != "" {
		if params.from, e = parseTime(raw); e != nil {
			return params, fmt.Errorf("from: %w", e)
		}
	} else {
		params.from = time.Now().UTC().Truncate(24 * time.Hour)
	}

	if raw := values.Get("to"); raw != "" {
		if params.to, e = parseTime(raw); e != nil {
			return params, fmt.Errorf("to: %w", e)
		}

		if !params.to.After(params.from) {
			return params, errors.New("to must be after from")
		}
	}

	return params, nil
}

func (stream *Stream) monitor(w http.ResponseWriter, r *http.Request) {
	connection, e := websocket.Accept(w, r, sameHost(r))

	if e != nil {
		return
	}

	defer connection.CloseNow()

	ctx := connection.CloseRead(r.Context())
	params, e := parseMonitor(r.URL.Query())

	if e != nil {
		send(ctx, connection, monitorMessage{Type: "error", Error: e.Error()})
		connection.Close(websocket.StatusPolicyViolation, e.Error())

		return
	}

	states := map[string]*stockState{}
	snapshot := monitorMessage{Type: "snapshot", Stocks: map[string]*stockUpdate{}}

	for _, stock := range params.stocks {
		state := &stockState{}
		states[stock] = state
		update, e := stream.load(state, stock, params, params.from)

		if e != nil {
			Logger.Error(StreamError, e)
		}

		snapshot.Stocks[stock] = update
	}

	if e := send(ctx, connection, snapshot); e != nil {
		return
	}

	if !params.to.IsZero() && params.to.Before(time.Now()) {
		<-ctx.Done()

		return
	}

	ticker := time.NewTicker(stream.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := monitorMessage{Type: "update", Stocks: map[string]*stockUpdate{}}

			for stock, state := range states {
				since := params.from

				if len(state.candles) > 0 {
					since = state.candles[len(state.candles)-1].Time
				}

				trading := state.contract != nil && state.contract.IsTradingHour()
				update, e := stream.load(state, stock, params, since)

				if e != nil {
					Logger.Error(StreamError, e)

					continue
				}

				if update.changed || update.Trading != trading {
					message.Stocks[stock] = update
				}
			}

			if len(message.Stocks) > 0 {
				if e := send(ctx, connection, message); e != nil {
					return
				}
			}

			if !params.to.IsZero() && time.Now().After(params.to) {
				<-ctx.Done()

				return
			}
		}
	}
}

func (stream *Stream) load(state *stockState, stock string, params monitorParams, since time.Time) (*stockUpdate, error) {
	update := &stockUpdate{Since: since, Candles: []Indicators.Candle{}, Signals: []Indicators.Signal{}}

	if time.Since(state.refreshed) >= CONTRACT_REFRESH {
		state.contract = stream.contract(stock)
		state.refreshed = time.Now()
	}

	update.Trading = state.contract != nil && state.contract.IsTradingHour()
	initial := len(state.candles) == 0
	warmup := 0

	if initial {
		warmup = params.warmup
	}

	fresh, e := Monitors.Candles(stream.db, stock, params.interval, since, params.to, warmup)

	if e != nil {
		return update, e
	}

	if !initial {
		cut := len(state.candles)

		for cut > 0 && !state.candles[cut-1].Time.Before(since) {
			cut--
		}

		if slices.Equal(state.candles[cut:], fresh) {
			return update, nil
		}

		state.candles = append(state.candles[:cut], fresh...)
	} else {
		state.candles = fresh
	}

	update.changed = true

	for _, candle := range state.candles {
		if !candle.Time.Before(since) && !candle.Time.Before(params.from) {
			update.Candles = append(update.Candles, candle)
		}
	}

	if len(update.Candles) == 0 && !initial {
		return update, nil
	}

	floor := since

	if floor.Before(params.from) {
		floor = params.from
	}

	for _, definition := range params.definitions {
		for _, signal := range definition.Evaluate(state.candles, floor) {
			signal.Expires = signal.Time.Add(params.lifetime)
			update.Signals = append(update.Signals, signal)
		}
	}

	return update, nil
}

func (stream *Stream) contract(stock string) *Market.Contract {
	contracts, e := latest[Tables.Contract](stream.db, Tables.CONTRACTS_TABLE, "stock_id", stock)

	if e != nil || len(contracts) == 0 {
		return nil
	}

	return contracts[0].Market(Tables.Stock{Id: stock})
}

func send(ctx context.Context, connection *websocket.Conn, message monitorMessage) error {
	raw, e := json.Marshal(message)

	if e != nil {
		return e
	}

	writeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return connection.Write(writeCtx, websocket.MessageText, raw)
}
