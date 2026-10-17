package Logger

import (
	"encoding/json"
	"strings"
	"time"
)

type Field int

const (
	FieldTime Field = iota
	FieldLevel
	FieldMessage
	FieldHost
	FieldFile
)

var DefaultFields = []Field{FieldTime, FieldLevel, FieldHost, FieldFile, FieldMessage}

const DefaultTimeLayout = time.RFC3339

type Formatter interface {
	Format(event Event) string
}

type TextFormatter struct {
	Fields     []Field
	TimeLayout string
	Separator  string
}

type JSONFormatter struct {
	Fields     []Field
	TimeLayout string
}

func (field Field) String() string {
	switch field {
	case FieldTime:
		return "time"
	case FieldLevel:
		return "level"
	case FieldMessage:
		return "message"
	case FieldHost:
		return "host"
	case FieldFile:
		return "file"
	}

	return "unknown"
}

func (field Field) value(event Event, timeLayout string) string {
	switch field {
	case FieldTime:
		if timeLayout == "" {
			timeLayout = DefaultTimeLayout
		}

		return event.Time.Format(timeLayout)
	case FieldLevel:
		return event.Level.String()
	case FieldMessage:
		return event.Message
	case FieldHost:
		return event.Host
	case FieldFile:
		return event.File
	}

	return ""
}

func fieldsOrDefault(fields []Field) []Field {
	if len(fields) == 0 {
		return DefaultFields
	}

	return fields
}

func (formatter TextFormatter) Format(event Event) string {
	separator := formatter.Separator

	if separator == "" {
		separator = " | "
	}

	fields := fieldsOrDefault(formatter.Fields)
	values := make([]string, 0, len(fields))

	for _, field := range fields {
		values = append(values, field.value(event, formatter.TimeLayout))
	}

	line := strings.Join(values, separator)

	if event.Trace != "" {
		line += "\n" + event.Trace
	}

	return line
}

func (formatter JSONFormatter) Format(event Event) string {
	var builder strings.Builder

	builder.WriteByte('{')

	for index, field := range fieldsOrDefault(formatter.Fields) {
		if index > 0 {
			builder.WriteByte(',')
		}

		writeJSONPair(&builder, field.String(), field.value(event, formatter.TimeLayout))
	}

	if event.Trace != "" {
		builder.WriteByte(',')
		writeJSONPair(&builder, "trace", event.Trace)
	}

	builder.WriteByte('}')

	return builder.String()
}

func writeJSONPair(builder *strings.Builder, key string, value string) {
	encodedKey, _ := json.Marshal(key)
	encodedValue, _ := json.Marshal(value)

	builder.Write(encodedKey)
	builder.WriteByte(':')
	builder.Write(encodedValue)
}
