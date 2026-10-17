package Control

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

const PATH = "/control"

type Handler func(ctx context.Context, message Message) Message

func Serve(handler Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+PATH, func(w http.ResponseWriter, r *http.Request) {
		connection, e := websocket.Accept(w, r, nil)

		if e != nil {
			return
		}

		defer connection.CloseNow()

		ctx := r.Context()
		writing := &sync.Mutex{}
		pending := &sync.WaitGroup{}

		defer pending.Wait()

		for {
			_, raw, e := connection.Read(ctx)

			if e != nil {
				return
			}

			message := Message{}

			if e := json.Unmarshal(raw, &message); e != nil {
				continue
			}

			pending.Add(1)

			go func(message Message) {
				defer pending.Done()

				reply, e := json.Marshal(handler(ctx, message))

				if e != nil {
					return
				}

				writing.Lock()
				defer writing.Unlock()

				connection.Write(ctx, websocket.MessageText, reply)
			}(message)
		}
	})

	return mux
}
