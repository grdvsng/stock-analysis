package Storage

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const (
	DB_TAG           = "db"
	DB_TYPE_TAG      = "dbtype"
	DESIGNATED_FLAG  = "designated"
	DEDUP_FLAG       = "dedup"
	DEFAULT_DECIMAL  = "DECIMAL(24,8)"
	DEFAULT_PARTITON = DAY_PARTITION
)

type Partition string

const (
	HOUR_PARTITION  Partition = "HOUR"
	DAY_PARTITION   Partition = "DAY"
	MONTH_PARTITION Partition = "MONTH"
	YEAR_PARTITION  Partition = "YEAR"
)

type Model interface {
	Table() string
}

type Partitioned interface {
	Partition() Partition
}

type field struct {
	index      []int
	column     string
	json       string
	sqlType    string
	designated bool
	dedup      bool
	encoded    bool
	textual    bool
}

type schema struct {
	fields     []field
	byColumn   map[string]*field
	designated *field
}

var (
	timeType    = reflect.TypeFor[time.Time]()
	decimalType = reflect.TypeFor[decimal.Decimal]()
	schemas     = sync.Map{}
)

func partition(model Model) Partition {
	if partitioned, ok := model.(Partitioned); ok {
		return partitioned.Partition()
	}

	return DEFAULT_PARTITON
}

func schemaOf(t reflect.Type) (*schema, error) {
	if cached, ok := schemas.Load(t); ok {
		return cached.(*schema), nil
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Model %s must be a struct", t)
	}

	result := &schema{byColumn: map[string]*field{}}

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		tag, tagged := structField.Tag.Lookup(DB_TAG)

		if !tagged || tag == "-" || !structField.IsExported() {
			continue
		}

		options := strings.Split(tag, ",")
		current := field{
			index:  structField.Index,
			column: options[0],
			json:   jsonName(structField),
		}

		for _, option := range options[1:] {
			switch option {
			case DESIGNATED_FLAG:
				current.designated = true
			case DEDUP_FLAG:
				current.dedup = true
			default:
				return nil, fmt.Errorf("Unknown option %q of %s.%s", option, t, structField.Name)
			}
		}

		current.sqlType, current.encoded = sqlType(structField.Type)
		current.textual = current.encoded && isTextual(structField.Type)
		current.encoded = current.encoded && !current.textual

		if custom, ok := structField.Tag.Lookup(DB_TYPE_TAG); ok {
			current.sqlType = custom
		}

		result.fields = append(result.fields, current)
	}

	for i := range result.fields {
		current := &result.fields[i]
		result.byColumn[current.column] = current

		if current.designated {
			if result.designated != nil {
				return nil, fmt.Errorf("Model %s has more than one designated timestamp", t)
			}

			if current.sqlType != "TIMESTAMP" {
				return nil, fmt.Errorf("Designated column %s of %s must be TIMESTAMP", current.column, t)
			}

			result.designated = current
		}
	}

	if result.designated == nil {
		return nil, fmt.Errorf("Model %s has no designated timestamp", t)
	}

	schemas.Store(t, result)

	return result, nil
}

func jsonName(structField reflect.StructField) string {
	name, _, _ := strings.Cut(structField.Tag.Get("json"), ",")

	if name == "" || name == "-" {
		return structField.Name
	}

	return name
}

var textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()

func isTextual(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Implements(textMarshalerType) || reflect.PointerTo(t).Implements(textMarshalerType)
}

func sqlType(t reflect.Type) (string, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t {
	case timeType:
		return "TIMESTAMP", false
	case decimalType:
		return DEFAULT_DECIMAL, false
	}

	switch t.Kind() {
	case reflect.String:
		return "VARCHAR", false
	case reflect.Bool:
		return "BOOLEAN", false
	case reflect.Int8:
		return "BYTE", false
	case reflect.Int16:
		return "SHORT", false
	case reflect.Int32:
		return "INT", false
	case reflect.Int, reflect.Int64:
		return "LONG", false
	case reflect.Float32:
		return "FLOAT", false
	case reflect.Float64:
		return "DOUBLE", false
	}

	return "VARCHAR", true
}

func (current *field) literal(value reflect.Value) (string, error) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "NULL", nil
		}

		value = value.Elem()
	}

	if current.textual {
		copied := reflect.New(value.Type())
		copied.Elem().Set(value)
		raw, e := copied.Interface().(encoding.TextMarshaler).MarshalText()

		return text(string(raw)), e
	}

	if current.encoded {
		raw, e := json.Marshal(value.Interface())

		return text(string(raw)), e
	}

	switch value.Type() {
	case timeType:
		moment := value.Interface().(time.Time)

		if moment.IsZero() {
			return "NULL", nil
		}

		return timestamp(moment), nil
	case decimalType:
		return text(value.Interface().(decimal.Decimal).String()), nil
	}

	switch value.Kind() {
	case reflect.String:
		return text(value.String()), nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return fmt.Sprint(value.Interface()), nil
	}

	return "", fmt.Errorf("Unsupported type %s of column %s", value.Type(), current.column)
}
