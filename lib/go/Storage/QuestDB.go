package Storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Result struct {
	Columns []Column            `json:"columns"`
	Dataset [][]json.RawMessage `json:"dataset"`
	Error   string              `json:"error"`
}

type QuestDB struct {
	endpoint string
	client   *http.Client
}

func Connect(host string, port, timeout int) (*QuestDB, error) {
	db := &QuestDB{
		endpoint: fmt.Sprintf("http://%s:%d/exec", host, port),
		client:   &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}

	_, e := db.Exec("SELECT 1")

	return db, e
}

func (db *QuestDB) Exec(query string) (*Result, error) {
	response, e := db.client.Get(db.endpoint + "?" + url.Values{"query": {query}}.Encode())

	if e != nil {
		return nil, e
	}

	defer response.Body.Close()

	result := &Result{}

	if e = json.NewDecoder(response.Body).Decode(result); e != nil {
		return nil, fmt.Errorf("%s: %w", response.Status, e)
	}

	if result.Error != "" {
		return result, fmt.Errorf("%s: %s", result.Error, query)
	}

	return result, nil
}
