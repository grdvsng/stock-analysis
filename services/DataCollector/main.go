package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Control"
	"github.com/grdvsng/stock-analysis/DataCollector/Collector"
	"github.com/grdvsng/stock-analysis/Exceptions"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Process"
	"github.com/grdvsng/stock-analysis/Result"
	"github.com/grdvsng/stock-analysis/Sources"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

const (
	REFRESH_INTERVAL = time.Minute
	LOG_DATE_FORMAT  = "2006-01-02"
)

var MetricsLoadError = Exceptions.CreateError("Can't load metrics: %s")

var InvalidMetricError = Exceptions.CreateError("Skip metric: %s")

var StartError = Exceptions.CreateError("Can't start data collector: %s")

var UpdateConfigError = Exceptions.CreateError("New config failed, restoring %s: %s")

var ControlError = Exceptions.CreateError("Control server on %s failed: %s")

type request struct {
	message Control.Message
	reply   chan Control.Message
}

type pending struct {
	request request
	archive string
}

type control struct {
	updates chan request
	metrics chan request
}

type service struct {
	db        *Storage.QuestDB
	worker    interface{ Close() }
	collector *Collector.Collector
}

func configPath() string {
	configPath := flag.String("config-path", "", "Path to the config file")
	ramLimit := flag.Float64("ram-limit", 0.5, "Ram limit for service in GB")

	flag.Parse()

	debug.SetMemoryLimit(int64(*ramLimit * 1024 * 1024 * 1024))

	return *configPath
}

func logPath(path string, day time.Time) string {
	extension := filepath.Ext(path)

	return fmt.Sprintf("%s.%s%s", strings.TrimSuffix(path, extension), day.Format(LOG_DATE_FORMAT), extension)
}

func configureLogger(config *Config.Config, day time.Time) {
	path := ""

	if config.Logging.Path != "" {
		path = logPath(config.Logging.Path, day)

		Result.Check(os.MkdirAll(filepath.Dir(path), 0o755))
	}

	Logger.Configure(Logger.Options{
		Level: config.Logging.Level,
		Formatter: Logger.TextFormatter{
			Separator: "\t",
		},
		FilePath: path,
	})
}

func nextDay(now time.Time) time.Duration {
	year, month, day := now.Date()

	return time.Date(year, month, day+1, 0, 0, 0, 0, now.Location()).Sub(now)
}

func loadMetrics(db *Storage.QuestDB) ([]Config.Metric, bool) {
	metrics, invalid, e := Config.LoadMetrics(db)

	if e != nil {
		Logger.Error(MetricsLoadError, e)

		return nil, false
	}

	for _, reason := range invalid {
		Logger.Error(InvalidMetricError, reason)
	}

	return metrics, true
}

func start(config *Config.Config) (*service, error) {
	db, e := Storage.Connect(config.Storage.Host, config.Storage.Port, config.Storage.Timeout)

	if e != nil {
		return nil, e
	}

	if e := Tables.Create(db); e != nil {
		return nil, e
	}

	if e := Config.CreateMetrics(db); e != nil {
		return nil, e
	}

	sources := make([]Sources.Source, 0, config.Workers)

	for i := 0; i < config.Workers; i++ {
		source, e := config.Source.Gateway(i)

		if e != nil {
			for _, opened := range sources {
				opened.Close()
			}

			return nil, e
		}

		sources = append(sources, source)
	}

	worker := Sources.CreateWorker(sources)

	return &service{
		db:        db,
		worker:    worker,
		collector: Collector.New(db, worker.Sources()),
	}, nil
}

func (service *service) close() {
	service.collector.Close()
	service.worker.Close()
}

func (control *control) handle(ctx context.Context, message Control.Message) Control.Message {
	current := request{message: message, reply: make(chan Control.Message, 1)}
	queue := control.metrics

	switch message.Type {
	case Control.UPDATE_CONFIG:
		queue = control.updates
	case Control.METRIC_UPDATED, Control.METRIC_REMOVED:
		if message.Metric == "" {
			return message.Reply(errors.New("metric required"))
		}
	default:
		return message.Reply(fmt.Errorf("unknown message type %q", message.Type))
	}

	select {
	case queue <- current:
	case <-ctx.Done():
		return message.Reply(ctx.Err())
	}

	select {
	case reply := <-current.reply:
		return reply
	case <-ctx.Done():
		return message.Reply(ctx.Err())
	}
}

func update(path string, current request) (*pending, error) {
	if _, e := Config.Validate(current.message.Config, path); e != nil {
		return nil, e
	}

	archive, e := Config.Archive(path, time.Now())

	if e != nil {
		return nil, e
	}

	if e := Config.Replace(path, current.message.Config); e != nil {
		return nil, e
	}

	return &pending{request: current, archive: archive}, nil
}

func signalMetric(service *service, current request) {
	metrics, ok := loadMetrics(service.db)

	if !ok {
		current.reply <- current.message.Reply(errors.New("can't load metrics"))

		return
	}

	var e error

	if current.message.Type == Control.METRIC_REMOVED {
		e = service.collector.Remove(current.message.Metric, metrics)
	} else {
		service.collector.Resubscribe(current.message.Metric, metrics)
	}

	current.reply <- current.message.Reply(e)
}

func run(ctx context.Context, path string, control *control, waiting *pending) (bool, *pending) {
	now := time.Now()
	config, e := Config.Load(path)
	var running *service

	if e == nil {
		configureLogger(config, now)

		running, e = start(config)
	}

	if waiting != nil {
		if e != nil {
			Logger.Error(UpdateConfigError, waiting.archive, e)

			if restoreError := Config.Restore(path, waiting.archive); restoreError != nil {
				e = errors.Join(e, restoreError)
			}

			waiting.request.reply <- waiting.request.message.Reply(fmt.Errorf("data collector failed with new config, previous config restored: %w", e))

			return true, nil
		}

		waiting.request.reply <- waiting.request.message.Reply(nil)
	} else if e != nil {
		Logger.Throw(StartError, e)
	}

	defer Logger.Close()

	Logger.Info("Service running")
	defer Logger.Info("Service stop")
	defer running.close()

	if metrics, ok := loadMetrics(running.db); ok {
		running.collector.Sync(metrics)
	}

	running.collector.Cleanup()

	restart := time.NewTimer(nextDay(now))
	reload := time.NewTicker(config.MetricsInterval())
	refresh := time.NewTicker(REFRESH_INTERVAL)

	defer restart.Stop()
	defer reload.Stop()
	defer refresh.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, nil
		case <-restart.C:
			Logger.Info("Day changed, restarting")

			return true, nil
		case current := <-control.updates:
			next, e := update(path, current)

			if e != nil {
				current.reply <- current.message.Reply(e)

				continue
			}

			Logger.Info("Config updated, previous archived to %s, restarting", next.archive)

			return true, next
		case current := <-control.metrics:
			signalMetric(running, current)
		case <-reload.C:
			if metrics, ok := loadMetrics(running.db); ok {
				running.collector.Sync(metrics)
			}
		case <-refresh.C:
			running.collector.Refresh()
		}
	}
}

func main() {
	path := configPath()
	config := Result.Ok(Config.Load(path))

	ctx, stop := Process.Context()
	defer stop()

	control := &control{updates: make(chan request), metrics: make(chan request)}
	address := fmt.Sprintf("%s:%d", config.Ws.Host, config.Ws.Port)
	server := &http.Server{Addr: address, Handler: Control.Serve(control.handle), ReadHeaderTimeout: 10 * time.Second}

	go func() {
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			Logger.Error(ControlError, address, e)
		}
	}()

	defer server.Shutdown(context.Background())

	var waiting *pending

	for {
		restart, next := run(ctx, path, control, waiting)

		if !restart {
			return
		}

		waiting = next
	}
}
