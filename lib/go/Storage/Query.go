package Storage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	TIMESTAMP_FORMAT  = "2006-01-02T15:04:05.000000Z"
	INSERT_BATCH_SIZE = 100
)

func text(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func timestamp(value time.Time) string {
	return text(value.UTC().Format(TIMESTAMP_FORMAT))
}

func Text(value string) string {
	return text(value)
}

func Timestamp(value time.Time) string {
	return timestamp(value)
}

func (db *QuestDB) Create(model Model) error {
	schema, e := schemaOf(reflect.TypeOf(model))

	if e != nil {
		return e
	}

	columns := make([]string, 0, len(schema.fields))
	dedup := []string{}

	for _, current := range schema.fields {
		columns = append(columns, current.column+" "+current.sqlType)

		if current.dedup && !current.designated {
			dedup = append(dedup, current.column)
		}
	}

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s) TIMESTAMP(%s) PARTITION BY %s WAL",
		model.Table(),
		strings.Join(columns, ", "),
		schema.designated.column,
		partition(model),
	)

	if len(dedup) > 0 || schema.designated.dedup {
		query += fmt.Sprintf(" DEDUP UPSERT KEYS(%s)", strings.Join(append([]string{schema.designated.column}, dedup...), ", "))
	}

	if _, e = db.Exec(query); e != nil {
		return e
	}

	return db.migrate(model.Table(), schema)
}

func (db *QuestDB) migrate(table string, schema *schema) error {
	result, e := db.Exec(fmt.Sprintf("SELECT \"column\" FROM table_columns(%s)", text(table)))

	if e != nil {
		return e
	}

	existing := map[string]bool{}

	for _, row := range result.Dataset {
		var column string

		if len(row) > 0 && json.Unmarshal(row[0], &column) == nil {
			existing[column] = true
		}
	}

	for _, current := range schema.fields {
		if existing[current.column] {
			continue
		}

		if _, e := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", table, current.column, current.sqlType)); e != nil {
			return e
		}
	}

	return nil
}

func (db *QuestDB) Insert(models ...Model) error {
	order := []string{}
	groups := map[string][]Model{}

	for _, model := range models {
		table := model.Table()

		if _, exists := groups[table]; !exists {
			order = append(order, table)
		}

		groups[table] = append(groups[table], model)
	}

	for _, table := range order {
		group := groups[table]

		for start := 0; start < len(group); start += INSERT_BATCH_SIZE {
			if e := db.insert(table, group[start:min(start+INSERT_BATCH_SIZE, len(group))]); e != nil {
				return e
			}
		}
	}

	return nil
}

func (db *QuestDB) insert(table string, models []Model) error {
	schema, e := schemaOf(reflect.TypeOf(models[0]))

	if e != nil {
		return e
	}

	columns := make([]string, 0, len(schema.fields))

	for _, current := range schema.fields {
		columns = append(columns, current.column)
	}

	rows := make([]string, 0, len(models))

	for _, model := range models {
		value := reflect.ValueOf(model)

		if value.Type() != reflect.TypeOf(models[0]) {
			return fmt.Errorf("Table %s gets mixed models %s and %s", table, value.Type(), reflect.TypeOf(models[0]))
		}

		literals := make([]string, 0, len(schema.fields))

		for i := range schema.fields {
			literal, e := schema.fields[i].literal(value.FieldByIndex(schema.fields[i].index))

			if e != nil {
				return e
			}

			literals = append(literals, literal)
		}

		rows = append(rows, "("+strings.Join(literals, ", ")+")")
	}

	_, e = db.Exec(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(rows, ", ")))

	return e
}

func Select[T Model](db *QuestDB, query string) ([]T, error) {
	schema, e := schemaOf(reflect.TypeFor[T]())

	if e != nil {
		return nil, e
	}

	result, e := db.Exec(query)

	if e != nil {
		return nil, e
	}

	records := make([]T, 0, len(result.Dataset))

	for _, row := range result.Dataset {
		object := make(map[string]json.RawMessage, len(row))

		for i, column := range result.Columns {
			current, exists := schema.byColumn[column.Name]

			if !exists || i >= len(row) {
				continue
			}

			raw := row[i]

			if current.encoded && string(raw) != "null" {
				var encoded string

				if e := json.Unmarshal(raw, &encoded); e != nil {
					return nil, e
				}

				raw = json.RawMessage(encoded)
			}

			object[current.json] = raw
		}

		data, e := json.Marshal(object)

		if e != nil {
			return nil, e
		}

		var record T

		if e := json.Unmarshal(data, &record); e != nil {
			return nil, e
		}

		records = append(records, record)
	}

	return records, nil
}

func (db *QuestDB) Exists(model Model) (bool, error) {
	result, e := db.Exec(fmt.Sprintf("SELECT count() FROM tables() WHERE table_name = %s", text(model.Table())))

	if e != nil {
		return false, e
	}

	count, e := scalar(result)

	return count > 0, e
}

func (db *QuestDB) Count(model Model) (int64, error) {
	result, e := db.Exec(fmt.Sprintf("SELECT count() FROM %s", model.Table()))

	if e != nil {
		return 0, e
	}

	return scalar(result)
}

func scalar(result *Result) (int64, error) {
	var value int64

	if len(result.Dataset) == 0 || len(result.Dataset[0]) == 0 {
		return 0, fmt.Errorf("Empty result")
	}

	return value, json.Unmarshal(result.Dataset[0][0], &value)
}

func (db *QuestDB) Drop(table string) error {
	_, e := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))

	return e
}

func (db *QuestDB) DropPartitionsBefore(model Model, before time.Time) error {
	schema, e := schemaOf(reflect.TypeOf(model))

	if e != nil {
		return e
	}

	_, e = db.Exec(fmt.Sprintf(
		"ALTER TABLE %s DROP PARTITION WHERE %s < timestamp_floor('%s', %s)",
		model.Table(),
		schema.designated.column,
		floorUnit(partition(model)),
		timestamp(before),
	))

	return e
}

func (db *QuestDB) Tables(prefix string) ([]string, error) {
	result, e := db.Exec(fmt.Sprintf("SELECT table_name FROM tables() WHERE table_name LIKE %s", text(prefix+"%")))

	if e != nil {
		return nil, e
	}

	tables := make([]string, 0, len(result.Dataset))

	for _, row := range result.Dataset {
		var table string

		if len(row) > 0 && json.Unmarshal(row[0], &table) == nil && strings.HasPrefix(table, prefix) {
			tables = append(tables, table)
		}
	}

	return tables, nil
}

func floorUnit(value Partition) string {
	switch value {
	case HOUR_PARTITION:
		return "h"
	case MONTH_PARTITION:
		return "M"
	case YEAR_PARTITION:
		return "y"
	}

	return "d"
}
