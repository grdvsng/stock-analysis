package Control

import "encoding/json"

type Type string

const (
	UPDATE_CONFIG  Type = "update_config"
	METRIC_UPDATED Type = "metric_updated"
	METRIC_REMOVED Type = "metric_removed"
	RESPONSE       Type = "response"
)

type Message struct {
	Id     string          `json:"id"`
	Type   Type            `json:"type"`
	Metric string          `json:"metric,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
	Ok     bool            `json:"ok"`
	Error  string          `json:"error,omitempty"`
}

func (message Message) Reply(e error) Message {
	reply := Message{Id: message.Id, Type: RESPONSE, Ok: e == nil}

	if e != nil {
		reply.Error = e.Error()
	}

	return reply
}
