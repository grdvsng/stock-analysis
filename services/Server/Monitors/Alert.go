package Monitors

import (
	"errors"
	"fmt"
	"strings"
)

type Condition string

const (
	BUY_CONDITION  Condition = "buy"
	SELL_CONDITION Condition = "sell"
	ANY_CONDITION  Condition = "any"
)

type Alert struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Stock     string    `json:"stock"`
	Metrics   []string  `json:"metrics"`
	Condition Condition `json:"condition"`
	Enabled   bool      `json:"enabled"`
}

func (alert *Alert) validate(stocks, notifications map[string]bool) error {
	alert.Name = strings.TrimSpace(alert.Name)

	if alert.Name == "" {
		return errors.New("alert name required")
	}

	if alert.Id == "" {
		alert.Id = NewId()
	}

	if alert.Condition == "" {
		alert.Condition = ANY_CONDITION
	}

	switch alert.Condition {
	case BUY_CONDITION, SELL_CONDITION, ANY_CONDITION:
	default:
		return fmt.Errorf("alert %s: unknown condition %q", alert.Name, alert.Condition)
	}

	if !stocks[alert.Stock] {
		return fmt.Errorf("alert %s: %s is not in the monitor", alert.Name, alert.Stock)
	}

	alert.Metrics = unique(alert.Metrics)

	if len(alert.Metrics) == 0 {
		return fmt.Errorf("alert %s: choose at least one metric", alert.Name)
	}

	for _, metric := range alert.Metrics {
		if !notifications[metric] {
			return fmt.Errorf("alert %s: %s is not a notification of the monitor", alert.Name, metric)
		}
	}

	return nil
}

func (monitor *Monitor) SanitizeAlerts() {
	stocks := set(monitor.Stocks)
	notifications := set(monitor.Notifications)
	kept := []Alert{}

	for _, alert := range monitor.Alerts {
		if !stocks[alert.Stock] {
			continue
		}

		metrics := []string{}

		for _, metric := range alert.Metrics {
			if notifications[metric] {
				metrics = append(metrics, metric)
			}
		}

		if len(metrics) > 0 {
			alert.Metrics = metrics
			kept = append(kept, alert)
		}
	}

	monitor.Alerts = kept
}

func (monitor *Monitor) validateAlerts() error {
	stocks := set(monitor.Stocks)
	notifications := set(monitor.Notifications)
	seen := map[string]bool{}

	for i := range monitor.Alerts {
		if e := monitor.Alerts[i].validate(stocks, notifications); e != nil {
			return e
		}

		if seen[monitor.Alerts[i].Id] {
			monitor.Alerts[i].Id = NewId()
		}

		seen[monitor.Alerts[i].Id] = true
	}

	return nil
}

func set(values []string) map[string]bool {
	result := make(map[string]bool, len(values))

	for _, value := range values {
		result[value] = true
	}

	return result
}
