package yugabyte

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type YugabyteDBStorage struct {
	db  *sql.DB
	sch *schema.Schema
}

func New() storage.Storage {
	return &YugabyteDBStorage{}
}

func (s *YugabyteDBStorage) Connect(ctx context.Context, connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *YugabyteDBStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.db == nil {
		return errors.New("not connected")
	}

	for _, table := range schema.Tables {
		createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY", table.TableName)

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}

			// Map data types to YugabyteDB equivalents (PostgreSQL-compatible)
			ybDataType := s.mapDataType(field.DataType)
			createTableQuery += fmt.Sprintf(", %s %s", field.Name, ybDataType)

			if field.Nullable {
				createTableQuery += " NULL"
			} else {
				createTableQuery += " NOT NULL"
			}
			if field.Default != nil {
				createTableQuery += fmt.Sprintf(" DEFAULT '%v'", field.Default)
			}
			if field.Unique {
				createTableQuery += " UNIQUE"
			}
		}

		createTableQuery += ")"

		storage.DebugLog(createTableQuery)
		log.Printf("Creating table %s with query: %s", table.TableName, createTableQuery)

		_, err := s.db.ExecContext(ctx, createTableQuery)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}
	}

	s.sch = schema
	log.Println("Tables created successfully")

	return nil
}

func (s *YugabyteDBStorage) mapDataType(dataType string) string {
	switch strings.ToUpper(dataType) {
	case "TEXT", "VARCHAR", "CHAR":
		return "TEXT"
	case "INTEGER", "INT":
		return "INTEGER"
	case "REAL", "FLOAT":
		return "REAL"
	case "BLOB":
		return "BYTEA"
	case "BOOLEAN":
		return "BOOLEAN"
	case "JSON":
		return "JSONB"
	case "DATETIME", "TIMESTAMP":
		return "TIMESTAMP"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIME"
	case "UUID":
		return "UUID"
	default:
		return "TEXT"
	}
}

func (s *YugabyteDBStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	if s.db == nil {
		return nil, false, errors.New("not connected")
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	if s.sch == nil {
		return nil, false, errors.New("schema not initialized")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return nil, false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	columns := make([]string, 0, len(obj.Fields)+1)
	placeholders := make([]string, 0, len(obj.Fields)+1)
	values := make([]any, 0, len(obj.Fields)+1)

	columns = append(columns, "id")
	placeholders = append(placeholders, "$1")
	values = append(values, obj.ID)

	paramIndex := 2
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))

		schField, ok := tbl.Fields[name]
		if !ok {
			return nil, false, fmt.Errorf("field %s not found in table %s schema", name, tbl.TableName)
		}
		if strings.ToLower(schField.DataType) == "json" {
			jsonData, err := json.Marshal(field)
			if err != nil {
				return nil, false, fmt.Errorf("failed to marshal JSON field %s: %w", name, err)
			}
			values = append(values, string(jsonData))
		} else {
			values = append(values, field)
		}
		paramIndex++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s",
		tbl.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		s.buildUpdateClause(columns, placeholders))

	storage.DebugLog(query, values...)
	_, err = s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *YugabyteDBStorage) buildUpdateClause(columns, placeholders []string) string {
	updateClauses := make([]string, 0, len(columns)-1)
	for i := 1; i < len(columns); i++ { // Skip id column
		updateClauses = append(updateClauses, fmt.Sprintf("%s = %s", columns[i], placeholders[i]))
	}
	return strings.Join(updateClauses, ", ")
}

func (s *YugabyteDBStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	setClauses := make([]string, 0, len(obj.Fields)-1)
	values := make([]any, 0, len(obj.Fields))
	paramIndex := 1

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", name, paramIndex))
		values = append(values, value)
		paramIndex++
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tbl.TableName, strings.Join(setClauses, ", "), paramIndex)
	storage.DebugLog(query, values...)

	res, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// Upsert inserts or updates an object, delegating to Insert which already implements upsert behavior using INSERT ... ON CONFLICT DO UPDATE
func (s *YugabyteDBStorage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return s.Insert(ctx, obj)
}

// UpsertTx inserts or updates an object within a transaction, delegating to InsertTx which already implements upsert behavior using INSERT ... ON CONFLICT DO UPDATE
func (s *YugabyteDBStorage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return s.InsertTx(ctx, tx, obj)
}

func (s *YugabyteDBStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.FindByKey(ctx, tblName, "id", id)
}

func (s *YugabyteDBStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	if tblName == "" || key == "" || value == "" {
		return nil, errors.New("table name, key, and value must not be empty")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	columns := make([]string, 0, len(tbl.Fields))
	columnPointers := make([]any, 0, len(tbl.Fields))

	var obj object.Object
	obj.TableName = tbl.TableName
	obj.Fields = make(map[string]any)

	for _, field := range tbl.Fields {
		columns = append(columns, field.Name)

		log.Printf("Processing field: %s with data type: %s", field.Name, field.DataType)
		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = columnPointer
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = *columnPointer
			}
		case "INTEGER", "INT":
			if field.Nullable {
				columnPointer := new(*int64)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = columnPointer
			} else {
				columnPointer := new(int64)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = *columnPointer
			}
		case "REAL", "FLOAT":
			columnPointer := new(float64)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "BLOB":
			columnPointer := new([]byte)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "BOOLEAN":
			columnPointer := new(bool)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "JSON":
			columnPointer := new(string)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			columnPointer := new(string)
			columnPointers = append(columnPointers, &columnPointer)
		default:
			return nil, fmt.Errorf("unsupported data type %s for field %s", field.DataType, field.Name)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", strings.Join(columns, ", "), tbl.TableName, tbl.Fields[key].Name)

	storage.DebugLog(query, value)

	row := s.db.QueryRowContext(ctx, query, value)
	if err := row.Scan(columnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &obj, nil
}

func (s *YugabyteDBStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}
	query := `DELETE FROM ` + tblName + ` WHERE id = $1`
	storage.DebugLog(query, id)
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *YugabyteDBStorage) ResetConnection(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Transaction methods

func (s *YugabyteDBStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, errors.New("not connected")
	}
	return s.db.BeginTx(ctx, nil)
}

func (s *YugabyteDBStorage) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Commit()
}

func (s *YugabyteDBStorage) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Rollback()
}

func (s *YugabyteDBStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	if tx == nil {
		return nil, false, errors.New("transaction is nil")
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	if s.sch == nil {
		return nil, false, errors.New("schema not initialized")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return nil, false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	columns := make([]string, 0, len(obj.Fields)+1)
	placeholders := make([]string, 0, len(obj.Fields)+1)
	values := make([]any, 0, len(obj.Fields)+1)

	columns = append(columns, "id")
	placeholders = append(placeholders, "$1")
	values = append(values, obj.ID)

	paramIndex := 2
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))

		schField, ok := tbl.Fields[name]
		if !ok {
			return nil, false, fmt.Errorf("field %s not found in table %s schema", name, tbl.TableName)
		}
		if strings.ToLower(schField.DataType) == "json" {
			jsonData, err := json.Marshal(field)
			if err != nil {
				return nil, false, fmt.Errorf("failed to marshal JSON field %s: %w", name, err)
			}
			values = append(values, string(jsonData))
		} else {
			values = append(values, field)
		}
		paramIndex++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s",
		tbl.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		s.buildUpdateClause(columns, placeholders))

	storage.DebugLog(query, values...)
	_, err = tx.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *YugabyteDBStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	setClauses := make([]string, 0, len(obj.Fields)-1)
	values := make([]any, 0, len(obj.Fields))
	paramIndex := 1

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", name, paramIndex))
		values = append(values, value)
		paramIndex++
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tbl.TableName, strings.Join(setClauses, ", "), paramIndex)
	storage.DebugLog(query, values...)

	res, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (s *YugabyteDBStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return s.FindByKeyTx(ctx, tx, tblName, "id", id)
}

func (s *YugabyteDBStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tblName == "" || key == "" || value == "" {
		return nil, errors.New("table name, key, and value must not be empty")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	columns := make([]string, 0, len(tbl.Fields))
	columnPointers := make([]any, 0, len(tbl.Fields))

	var obj object.Object
	obj.TableName = tbl.TableName
	obj.Fields = make(map[string]any)

	for _, field := range tbl.Fields {
		columns = append(columns, field.Name)

		log.Printf("Processing field: %s with data type: %s", field.Name, field.DataType)
		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = columnPointer
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = *columnPointer
			}
		case "INTEGER", "INT":
			if field.Nullable {
				columnPointer := new(*int64)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = columnPointer
			} else {
				columnPointer := new(int64)
				columnPointers = append(columnPointers, &columnPointer)
				obj.Fields[field.Name] = *columnPointer
			}
		case "REAL", "FLOAT":
			columnPointer := new(float64)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "BLOB":
			columnPointer := new([]byte)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "BOOLEAN":
			columnPointer := new(bool)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "JSON":
			columnPointer := new(string)
			columnPointers = append(columnPointers, &columnPointer)
			obj.Fields[field.Name] = columnPointer
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			columnPointer := new(string)
			columnPointers = append(columnPointers, &columnPointer)
		default:
			return nil, fmt.Errorf("unsupported data type %s for field %s", field.DataType, field.Name)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", strings.Join(columns, ", "), tbl.TableName, tbl.Fields[key].Name)

	storage.DebugLog(query, value)

	row := tx.QueryRowContext(ctx, query, value)
	if err := row.Scan(columnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &obj, nil
}

func (s *YugabyteDBStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}
	query := `DELETE FROM ` + tblName + ` WHERE id = $1`
	storage.DebugLog(query, id)
	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
