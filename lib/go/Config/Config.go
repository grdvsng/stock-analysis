package Config

import (
	"os"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Time/Duration"
	"github.com/spf13/viper"
)

type Config struct {
	Source         source             `json:"source"`
	UpdateInterval *Duration.Duration `json:"update_interval"`
	Storage        storage            `json:"storage"`
	Logging        logging            `json:"logging"`
	Ws             ws                 `json:"ws"`
	Workers        int                `json:"workers"`
}

func (config *Config) MetricsInterval() time.Duration {
	return config.UpdateInterval.Duration()
}

func Load(path string) (*Config, error) {
	var config Config

	if path == "" {
		return nil, PathRequired("Config")
	} else if _, err := os.Stat(path); err != nil {
		return nil, PathDoesNotExists(path)
	}

	viper := viper.New()

	viper.SetTypeByDefaultValue(true)

	viper.SetDefault("workers", 4)
	viper.SetDefault("update_interval", "10 Seconds")
	viper.SetDefault("source.timeout", 300)
	viper.SetDefault("source.client_id", 1)
	viper.SetDefault("logging.level", Logger.INFO)
	viper.SetDefault("logging.path", "/var/log/stock-analysis/data-collector.log")
	viper.SetDefault("storage.host", "localhost")
	viper.SetDefault("storage.port", 9000)
	viper.SetDefault("storage.timeout", 30)
	viper.SetDefault("ws.host", "127.0.0.1")
	viper.SetDefault("ws.port", 9100)
	viper.SetConfigType("json")

	viper.SetConfigFile(path)

	e := viper.ReadInConfig()

	if e == nil {
		e = viper.Unmarshal(&config, func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
			dc.DecodeHook = mapstructure.TextUnmarshallerHookFunc()
		})
	}

	if e == nil {
		e = config.validate()
	}

	return &config, e
}

func (config *Config) validate() error {
	if config.Workers < 1 {
		return InvalidValue("workers", config.Workers)
	}

	if config.Ws.Port < 1 || config.Ws.Port > 65535 {
		return InvalidValue("ws.port", config.Ws.Port)
	}

	return nil
}
