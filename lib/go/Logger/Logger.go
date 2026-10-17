package Logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/grdvsng/stock-analysis/Exceptions"
)

type Options struct {
	Level     Level
	Formatter Formatter
	FilePath  string
}

type Logger struct {
	level     Level
	formatter Formatter
	host      string
	file      *os.File
	mutex     *sync.Mutex
}

func New(options Options) *Logger {
	host, _ := os.Hostname()
	logger := &Logger{
		mutex: &sync.Mutex{},
		host:  host,
	}

	logger.configure(options)

	return logger
}

func (logger *Logger) configure(options Options) {
	logger.level = options.Level
	logger.formatter = options.Formatter

	if logger.formatter == nil {
		logger.formatter = TextFormatter{}
	}

	if logger.file != nil {
		logger.file.Close()
		logger.file = nil
	}

	if options.FilePath != "" {
		file, err := os.OpenFile(options.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

		if err != nil {
			panic(err)
		}

		logger.file = file
	}
}

func (logger *Logger) skip(l Level) bool {
	return logger.level > l
}

func (logger *Logger) Close() error {
	if logger.file == nil {
		return nil
	}

	err := logger.file.Close()

	if errors.Is(err, os.ErrClosed) {
		return nil
	}

	return err
}

func (logger *Logger) Debug(message string, args ...any) {
	if logger.skip(DEBUG_LEVEL) {
		return
	}

	logger.log(DEBUG_LEVEL, message, args...)
}

func (logger *Logger) Info(message string, args ...any) {
	if logger.skip(INFO_LEVEL) {
		return
	}

	logger.log(INFO_LEVEL, message, args...)
}

func (logger *Logger) Warning(w Exceptions.Warning, args ...any) {
	if logger.skip(WARNING_LEVEL) {
		return
	}

	logger.log(WARNING_LEVEL, "%s", w.Render(true, args...))
}

func (logger *Logger) Error(err Exceptions.Error, args ...any) {
	if logger.skip(ERROR_LEVEL) {
		return
	}

	logger.log(ERROR_LEVEL, "%s", err.Render(true, args...))
}

func (logger *Logger) Critical(err Exceptions.Error, args ...any) {
	if logger.skip(CRITICAL_LEVEL) {
		return
	}

	logger.log(CRITICAL_LEVEL, err.Render(false), args...)
}

func (logger *Logger) Exception(err Exceptions.Error, args ...any) {
	if logger.skip(ERROR_LEVEL) {
		return
	}

	logger.log(ERROR_LEVEL, "%s", err.Render(true, args...))
}

func (logger *Logger) Throw(err Exceptions.Error, args ...any) {
	if !logger.skip(CRITICAL_LEVEL) {
		logger.Critical(err, args...)
	}

	logger.Close()
	err.Panic(args...)
}

func (logger *Logger) log(level Level, message string, args ...any) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	event := Event{
		Time:  time.Now(),
		Level: level,
		Host:  logger.host,
		File:  caller(2),
	}

	if len(args) == 0 {
		event.Message = message
	} else {
		event.Message = fmt.Sprintf(message, args...)
	}

	line := logger.formatter.Format(event) + "\n"

	os.Stdout.WriteString(line)

	if logger.file != nil {
		logger.file.WriteString(line)
	}
}

func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)

	if !ok {
		return ""
	}

	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
