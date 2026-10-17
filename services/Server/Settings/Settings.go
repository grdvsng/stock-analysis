package Settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Time/Duration"
)

const (
	DEFAULT_HTTP_PORT       = 8080
	DEFAULT_WS_PORT         = 8081
	DEFAULT_STREAM_INTERVAL = 2 * time.Second
	DEFAULT_UPDATE_TIMEOUT  = 10 * time.Minute
)

type logging struct {
	Level Logger.Level `json:"level"`
	Path  string       `json:"path"`
}

type Settings struct {
	HttpPort            int                `json:"http_port"`
	WsPort              int                `json:"ws_port"`
	DataCollectorConfig string             `json:"data_collector_config"`
	StreamInterval      *Duration.Duration `json:"stream_interval"`
	UpdateTimeout       *Duration.Duration `json:"update_timeout"`
	UiPath              string             `json:"ui_path"`
	Logging             logging            `json:"logging"`
}

func (settings *Settings) StreamEvery() time.Duration {
	if settings.StreamInterval != nil && settings.StreamInterval.Duration() > 0 {
		return settings.StreamInterval.Duration()
	}

	return DEFAULT_STREAM_INTERVAL
}

func (settings *Settings) UpdateWait() time.Duration {
	if settings.UpdateTimeout != nil && settings.UpdateTimeout.Duration() > 0 {
		return settings.UpdateTimeout.Duration()
	}

	return DEFAULT_UPDATE_TIMEOUT
}

func Load(path string) (*Settings, error) {
	if path == "" {
		return nil, fmt.Errorf("Server config path required.")
	}

	raw, e := os.ReadFile(path)

	if e != nil {
		return nil, e
	}

	settings := &Settings{
		HttpPort: DEFAULT_HTTP_PORT,
		WsPort:   DEFAULT_WS_PORT,
		Logging:  logging{Level: Logger.INFO_LEVEL},
	}

	if e := json.Unmarshal(raw, settings); e != nil {
		return nil, e
	}

	if settings.DataCollectorConfig == "" {
		return nil, fmt.Errorf("data_collector_config required.")
	}

	settings.DataCollectorConfig = relative(path, settings.DataCollectorConfig)

	if settings.UiPath != "" {
		settings.UiPath = relative(path, settings.UiPath)
	}

	return settings, nil
}

func relative(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(filepath.Dir(base), path)
}
