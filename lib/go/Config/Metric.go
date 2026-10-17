package Config

import (
	"fmt"
	"time"

	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Time/Duration"
	"github.com/grdvsng/stock-analysis/Time/TTL"
)

const (
	METRICS_TABLE                    = "metrics"
	DEFAULT_STOCK_UPDATE_INTERVAL    = 10 * time.Second
	DEFAULT_CONTRACT_UPDATE_INTERVAL = 24 * time.Hour
	DEFAULT_STOCK_TTL                = 30 * 24 * time.Hour
)

type Metric struct {
	ID                     string             `json:"id" db:"id,dedup" dbtype:"SYMBOL"`
	TTL                    *TTL.TTL           `json:"ttl" db:"ttl"`
	UpdateInterval         *Duration.Duration `json:"update_interval" db:"update_interval"`
	ContractUpdateInterval *Duration.Duration `json:"contract_update_interval" db:"contract_update_interval"`
	Reprocess              int                `json:"reprocess" db:"reprocess"`
	Enabled                bool               `json:"enabled" db:"enabled"`
	CreatedAt              time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at" db:"updated_at,designated"`
}

func (Metric) Table() string {
	return METRICS_TABLE
}

func (Metric) Partition() Storage.Partition {
	return Storage.YEAR_PARTITION
}

func (metric *Metric) Interval() time.Duration {
	return durationOr(metric.UpdateInterval, DEFAULT_STOCK_UPDATE_INTERVAL)
}

func (metric *Metric) ContractInterval() time.Duration {
	return durationOr(metric.ContractUpdateInterval, DEFAULT_CONTRACT_UPDATE_INTERVAL)
}

func (metric *Metric) Retention() time.Duration {
	if metric.TTL != nil && metric.TTL.Duration() > 0 {
		return metric.TTL.Duration()
	}

	return DEFAULT_STOCK_TTL
}

func (metric *Metric) Validate() error {
	if metric.ID == "" {
		return InvalidValue("metrics.id", metric.ID)
	}

	if metric.Reprocess < 0 {
		return InvalidValue("metrics.reprocess", metric.Reprocess)
	}

	if time.Duration(metric.Reprocess)*24*time.Hour > metric.Retention() {
		return ReprocessExceedsTTL(metric.ID, metric.Reprocess)
	}

	return nil
}

func CreateMetrics(db *Storage.QuestDB) error {
	return db.Create(Metric{})
}

func LoadMetrics(db *Storage.QuestDB) ([]Metric, []error, error) {
	if exists, e := db.Exists(Metric{}); e != nil || !exists {
		return nil, nil, e
	}

	latest, e := Storage.Select[Metric](db, fmt.Sprintf(
		"SELECT * FROM (SELECT * FROM %s LATEST ON updated_at PARTITION BY id) WHERE enabled",
		METRICS_TABLE,
	))

	if e != nil {
		return nil, nil, e
	}

	metrics := make([]Metric, 0, len(latest))
	invalid := []error{}

	for _, metric := range latest {
		if e := metric.Validate(); e != nil {
			invalid = append(invalid, e)
		} else {
			metrics = append(metrics, metric)
		}
	}

	return metrics, invalid, nil
}

func LoadMetric(db *Storage.QuestDB, id string) (*Metric, error) {
	if exists, e := db.Exists(Metric{}); e != nil || !exists {
		return nil, e
	}

	latest, e := Storage.Select[Metric](db, fmt.Sprintf(
		"SELECT * FROM %s WHERE id = %s LATEST ON updated_at PARTITION BY id",
		METRICS_TABLE,
		Storage.Text(id),
	))

	if e != nil || len(latest) == 0 || !latest[0].Enabled {
		return nil, e
	}

	return &latest[0], nil
}

func SaveMetric(db *Storage.QuestDB, metric Metric) error {
	if e := metric.Validate(); e != nil {
		return e
	}

	now := time.Now()
	metric.Enabled = true
	metric.UpdatedAt = now
	metric.CreatedAt = now

	if current, e := LoadMetric(db, metric.ID); e != nil {
		return e
	} else if current != nil && !current.CreatedAt.IsZero() {
		metric.CreatedAt = current.CreatedAt
	}

	return db.Insert(metric)
}

func RemoveMetric(db *Storage.QuestDB, id string) error {
	return db.Insert(Metric{ID: id, Enabled: false, UpdatedAt: time.Now()})
}

func durationOr(duration *Duration.Duration, fallback time.Duration) time.Duration {
	if duration != nil && duration.Duration() > 0 {
		return duration.Duration()
	}

	return fallback
}
