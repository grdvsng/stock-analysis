package Control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

const RECONNECT_INTERVAL = 5 * time.Second

var ErrDisconnected = errors.New("data collector is not connected")

type Client struct {
	url        string
	mutex      *sync.Mutex
	connection *websocket.Conn
	waiting    map[string]chan Message
	counter    *atomic.Int64
}

func NewClient(host string, port int) *Client {
	return &Client{
		url:     fmt.Sprintf("ws://%s:%d%s", host, port, PATH),
		mutex:   &sync.Mutex{},
		waiting: map[string]chan Message{},
		counter: &atomic.Int64{},
	}
}

func (client *Client) Connected() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	return client.connection != nil
}

func (client *Client) Run(ctx context.Context, onError func(error)) {
	for ctx.Err() == nil {
		connection, _, e := websocket.Dial(ctx, client.url, nil)

		if e == nil {
			connection.SetReadLimit(-1)
			client.mutex.Lock()
			client.connection = connection
			client.mutex.Unlock()

			e = client.read(ctx, connection)

			client.mutex.Lock()
			client.connection = nil

			for id, waiter := range client.waiting {
				close(waiter)
				delete(client.waiting, id)
			}

			client.mutex.Unlock()
			connection.CloseNow()
		}

		if ctx.Err() == nil && onError != nil {
			onError(e)
		}

		select {
		case <-ctx.Done():
		case <-time.After(RECONNECT_INTERVAL):
		}
	}
}

func (client *Client) read(ctx context.Context, connection *websocket.Conn) error {
	for {
		_, raw, e := connection.Read(ctx)

		if e != nil {
			return e
		}

		message := Message{}

		if json.Unmarshal(raw, &message) != nil {
			continue
		}

		client.mutex.Lock()
		waiter, exists := client.waiting[message.Id]
		delete(client.waiting, message.Id)
		client.mutex.Unlock()

		if exists {
			waiter <- message
			close(waiter)
		}
	}
}

func (client *Client) Request(ctx context.Context, message Message) (Message, error) {
	message.Id = strconv.FormatInt(client.counter.Add(1), 10)
	raw, e := json.Marshal(message)

	if e != nil {
		return Message{}, e
	}

	waiter := make(chan Message, 1)

	client.mutex.Lock()
	connection := client.connection

	if connection == nil {
		client.mutex.Unlock()

		return Message{}, ErrDisconnected
	}

	client.waiting[message.Id] = waiter
	client.mutex.Unlock()

	if e := connection.Write(ctx, websocket.MessageText, raw); e != nil {
		client.forget(message.Id)

		return Message{}, e
	}

	select {
	case <-ctx.Done():
		client.forget(message.Id)

		return Message{}, ctx.Err()
	case reply, open := <-waiter:
		if !open {
			return Message{}, ErrDisconnected
		}

		return reply, nil
	}
}

func (client *Client) forget(id string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.waiting, id)
}
