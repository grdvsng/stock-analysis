package Api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

const CLIENT_BUFFER = 16

type streamMessage struct {
	Metrics map[string][]Tables.Metric `json:"metrics"`
}

type Stream struct {
	db        *Storage.QuestDB
	interval  time.Duration
	mutex     *sync.Mutex
	clients   map[chan []byte]bool
	since     time.Time
	watermark map[string]time.Time
}

func NewStream(db *Storage.QuestDB, interval time.Duration) *Stream {
	return &Stream{
		db:        db,
		interval:  interval,
		mutex:     &sync.Mutex{},
		clients:   map[chan []byte]bool{},
		since:     time.Now().UTC(),
		watermark: map[string]time.Time{},
	}
}

func (stream *Stream) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /contracts", stream.serve)
	mux.HandleFunc("GET /monitor", stream.monitor)

	return mux
}

func sameHost(r *http.Request) *websocket.AcceptOptions {
	host := r.Host

	if parsed, _, e := net.SplitHostPort(r.Host); e == nil {
		host = parsed
	}

	return &websocket.AcceptOptions{OriginPatterns: []string{host, host + ":*"}}
}

func (stream *Stream) Run(ctx context.Context) {
	ticker := time.NewTicker(stream.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e := stream.publish(); e != nil {
				Logger.Error(StreamError, e)
			}
		}
	}
}

func (stream *Stream) serve(w http.ResponseWriter, r *http.Request) {
	connection, e := websocket.Accept(w, r, sameHost(r))

	if e != nil {
		return
	}

	defer connection.CloseNow()

	messages := make(chan []byte, CLIENT_BUFFER)

	stream.mutex.Lock()
	stream.clients[messages] = true
	stream.mutex.Unlock()

	defer func() {
		stream.mutex.Lock()
		delete(stream.clients, messages)
		stream.mutex.Unlock()
	}()

	ctx := connection.CloseRead(r.Context())

	for {
		select {
		case <-ctx.Done():
			return
		case message, open := <-messages:
			if !open {
				connection.Close(websocket.StatusPolicyViolation, "client is too slow")

				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, stream.interval*5)
			e := connection.Write(writeCtx, websocket.MessageText, message)
			cancel()

			if e != nil {
				return
			}
		}
	}
}

func (stream *Stream) publish() error {
	stream.mutex.Lock()
	listeners := len(stream.clients)
	stream.mutex.Unlock()

	metrics, e := stream.collect()

	if e != nil || len(metrics) == 0 || listeners == 0 {
		return e
	}

	message, e := json.Marshal(streamMessage{Metrics: metrics})

	if e != nil {
		return e
	}

	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	for client := range stream.clients {
		select {
		case client <- message:
		default:
			delete(stream.clients, client)
			close(client)
		}
	}

	return nil
}

func (stream *Stream) collect() (map[string][]Tables.Metric, error) {
	tables, e := stream.db.Tables(Tables.METRICS_TABLE_PREFIX)

	if e != nil {
		return nil, e
	}

	result := map[string][]Tables.Metric{}

	for _, table := range tables {
		since, known := stream.watermark[table]

		if !known {
			since = stream.since
		}

		rows, e := Storage.Select[Tables.Metric](stream.db, fmt.Sprintf(
			"SELECT * FROM %s WHERE time > %s ORDER BY time",
			table,
			Storage.Timestamp(since),
		))

		if e != nil {
			return nil, e
		}

		if len(rows) > 0 {
			since = rows[len(rows)-1].Time
		}

		stream.watermark[table] = since

		for _, row := range rows {
			result[row.StockId] = append(result[row.StockId], row)
		}
	}

	return result, nil
}
