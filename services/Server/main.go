package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grdvsng/stock-analysis/Config"
	"github.com/grdvsng/stock-analysis/Control"
	"github.com/grdvsng/stock-analysis/Exceptions"
	"github.com/grdvsng/stock-analysis/Logger"
	"github.com/grdvsng/stock-analysis/Process"
	"github.com/grdvsng/stock-analysis/Result"
	"github.com/grdvsng/stock-analysis/Server/Api"
	"github.com/grdvsng/stock-analysis/Server/Monitors"
	"github.com/grdvsng/stock-analysis/Server/Settings"
	"github.com/grdvsng/stock-analysis/Storage"
	"github.com/grdvsng/stock-analysis/Storage/Tables"
)

const SHUTDOWN_TIMEOUT = 10 * time.Second

var ListenError = Exceptions.CreateError("%s server on port %d failed: %s")

var ControlWarning = Exceptions.CreateWarning("Data collector control %s:%d unavailable: %s")

func configureLogger(settings *Settings.Settings) {
	if settings.Logging.Path != "" {
		Result.Check(os.MkdirAll(filepath.Dir(settings.Logging.Path), 0o755))
	}

	Logger.Configure(Logger.Options{
		Level: settings.Logging.Level,
		Formatter: Logger.TextFormatter{
			Separator: "\t",
		},
		FilePath: settings.Logging.Path,
	})
}

func listen(name string, port int, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		Logger.Info("%s server listening on %d", name, port)

		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			Logger.Error(ListenError, name, port, e)
		}
	}()

	return server
}

func main() {
	configPath := flag.String("config-path", "", "Path to the server config file")

	flag.Parse()

	settings := Result.Ok(Settings.Load(*configPath))

	configureLogger(settings)
	defer Logger.Close()

	collector := Result.Ok(Config.Load(settings.DataCollectorConfig))
	db := Result.Ok(Storage.Connect(collector.Storage.Host, collector.Storage.Port, collector.Storage.Timeout))

	Result.Check(Tables.Create(db))
	Result.Check(Config.CreateMetrics(db))
	Result.Check(Monitors.Create(db))

	ctx, stop := Process.Context()
	defer stop()

	control := Control.NewClient(collector.Ws.Host, collector.Ws.Port)

	go control.Run(ctx, func(e error) {
		Logger.Warning(ControlWarning, collector.Ws.Host, collector.Ws.Port, e)
	})

	api := Api.New(db, settings.DataCollectorConfig, collector, control, settings.UpdateWait(), settings.UiPath, settings.WsPort)
	defer api.Close()

	stream := Api.NewStream(db, settings.StreamEvery())

	go stream.Run(ctx)

	httpServer := listen("HTTP", settings.HttpPort, api.Handler())
	wsServer := listen("WS", settings.WsPort, stream.Handler())

	Logger.Info("Server running")

	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), SHUTDOWN_TIMEOUT)
	defer cancel()

	httpServer.Shutdown(shutdown)
	wsServer.Shutdown(shutdown)

	Logger.Info("Server stop")
}
