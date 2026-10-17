package Monitors

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Indicators"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Time/Duration"
)

const MONITORS_TABLE = "monitors"

type Monitor struct {
	Id            string             `json:"id" db:"id,dedup" dbtype:"SYMBOL"`
	Name          string             `json:"name" db:"name"`
	Stocks        []string           `json:"stocks" db:"stocks"`
	Notifications []string           `json:"notifications" db:"notifications"`
	Interval      string             `json:"interval" db:"interval" dbtype:"SYMBOL"`
	EventLifetime *Duration.Duration `json:"event_lifetime" db:"event_lifetime"`
	Alerts        []Alert            `json:"alerts" db:"alerts"`
	Enabled       bool               `json:"enabled" db:"enabled"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" db:"updated_at,designated"`
}

func (Monitor) Table() string {
	return MONITORS_TABLE
}

func (Monitor) Partition() Storage.Partition {
	return Storage.YEAR_PARTITION
}

func (monitor *Monitor) Validate() error {
	monitor.Name = strings.TrimSpace(monitor.Name)

	if monitor.Name == "" {
		return errors.New("name required")
	}

	if _, exists := Intervals[monitor.Interval]; !exists {
		return fmt.Errorf("unknown interval %q", monitor.Interval)
	}

	if monitor.EventLifetime != nil && monitor.EventLifetime.Duration() <= 0 {
		return errors.New("event lifetime must be positive")
	}

	monitor.Stocks = unique(monitor.Stocks)
	monitor.Notifications = unique(monitor.Notifications)

	for _, key := range monitor.Notifications {
		if _, exists := Indicators.Lookup(key); !exists {
			return fmt.Errorf("unknown notification %q", key)
		}
	}

	if monitor.Alerts == nil {
		monitor.Alerts = []Alert{}
	}

	return monitor.validateAlerts()
}

func (monitor *Monitor) Lifetime() time.Duration {
	return Lifetime(monitor.EventLifetime, monitor.Interval)
}

func Lifetime(lifetime *Duration.Duration, interval string) time.Duration {
	if lifetime != nil && lifetime.Duration() > 0 {
		return lifetime.Duration()
	}

	return Intervals[interval]
}

func unique(values []string) []string {
	seen := map[string]bool{}
	result := []string{}

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	return result
}

func NewId() string {
	raw := make([]byte, 8)
	rand.Read(raw)

	return hex.EncodeToString(raw)
}

func Create(db *Storage.QuestDB) error {
	return db.Create(Monitor{})
}

func List(db *Storage.QuestDB) ([]Monitor, error) {
	return Storage.Select[Monitor](db, fmt.Sprintf(
		"SELECT * FROM (SELECT * FROM %s LATEST ON updated_at PARTITION BY id) WHERE enabled ORDER BY created_at",
		MONITORS_TABLE,
	))
}

func Get(db *Storage.QuestDB, id string) (*Monitor, error) {
	latest, e := Storage.Select[Monitor](db, fmt.Sprintf(
		"SELECT * FROM %s WHERE id = %s LATEST ON updated_at PARTITION BY id",
		MONITORS_TABLE,
		Storage.Text(id),
	))

	if e != nil || len(latest) == 0 || !latest[0].Enabled {
		return nil, e
	}

	return &latest[0], nil
}

func Save(db *Storage.QuestDB, monitor Monitor) (Monitor, error) {
	if e := monitor.Validate(); e != nil {
		return monitor, e
	}

	now := time.Now().UTC()

	if monitor.CreatedAt.IsZero() {
		monitor.CreatedAt = now
	}

	monitor.Enabled = true
	monitor.UpdatedAt = now

	return monitor, db.Insert(monitor)
}

func Remove(db *Storage.QuestDB, id string) error {
	return db.Insert(Monitor{Id: id, Interval: DEFAULT_INTERVAL, Enabled: false, UpdatedAt: time.Now().UTC()})
}
