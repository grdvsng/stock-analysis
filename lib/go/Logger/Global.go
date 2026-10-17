package Logger

import (
	"sync"

	"github.com/grdvsng/stock-analysis/Exceptions"
)

var m = &sync.Mutex{}
var instance = New(Options{Level: INFO_LEVEL})

func Configure(options Options) *Logger {
	m.Lock()
	defer m.Unlock()

	instance.configure(options)

	return instance
}

func Close() error {
	m.Lock()
	defer m.Unlock()

	return instance.Close()
}

func Debug(message string, args ...any) {
	instance.Debug(message, args...)
}

func Info(message string, args ...any) {
	instance.Info(message, args...)
}

func Warning(w Exceptions.Warning, args ...any) {
	instance.Warning(w, args...)
}

func Error(e Exceptions.Error, args ...any) {
	instance.Error(e, args...)
}

func Critical(e Exceptions.Error, args ...any) {
	instance.Critical(e, args...)
}

func Exception(err Exceptions.Error, args ...any) {
	instance.Exception(err, args...)
}

func Throw(err Exceptions.Error, args ...any) {
	instance.Throw(err, args...)
}
