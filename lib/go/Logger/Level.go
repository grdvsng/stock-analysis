package Logger

import "fmt"

type Level int

const (
	DEBUG_LEVEL Level = iota
	INFO_LEVEL
	WARNING_LEVEL
	ERROR_LEVEL
	CRITICAL_LEVEL
	DISABLED_LEVEL
)

const (
	DEBUG    = "DEBUG"
	INFO     = "INFO"
	WARNING  = "WARNING"
	ERROR    = "ERROR"
	CRITICAL = "CRITICAL"
	UNKNOWN  = "UNKNOWN"
	DISABLED = "DISABLED"
)

func (level Level) String() string {
	switch level {
	case DEBUG_LEVEL:
		return DEBUG
	case INFO_LEVEL:
		return INFO
	case WARNING_LEVEL:
		return WARNING
	case ERROR_LEVEL:
		return ERROR
	case CRITICAL_LEVEL:
		return CRITICAL
	case DISABLED_LEVEL:
		return DISABLED
	}

	return UNKNOWN
}

func (level *Level) UnmarshalText(b []byte) (err error) {
	value := string(b)

	switch value {
	case DEBUG:
		*level = DEBUG_LEVEL
	case INFO:
		*level = INFO_LEVEL
	case WARNING:
		*level = WARNING_LEVEL
	case ERROR:
		*level = ERROR_LEVEL
	case CRITICAL:
		*level = CRITICAL_LEVEL
	case DISABLED:
		*level = DISABLED_LEVEL
	default:
		return fmt.Errorf("Undefined log level %s", value)
	}

	return nil
}
