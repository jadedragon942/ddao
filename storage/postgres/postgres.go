package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
	_ "github.com/lib/pq"
)

type PostgreSQLStorage struct {
	db  *sql.DB
	sch *schema.Schema
	mu  *sync.Mutex
}

func New() storage.Storage {
	var mu sync.Mutex
	return &PostgreSQLStorage{mu: &mu}
}

func (s *PostgreSQLStorage) Connect(ctx context.Context, connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *PostgreSQLStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.db == nil {
		return errors.New("not connected")
	}

	for _, table := range schema.Tables {
		createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY", table.TableName)

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}

			// Map data types to PostgreSQL equivalents
			pgDataType := s.mapDataType(field.DataType)
			createTableQuery += fmt.Sprintf(", %s %s", field.Name, pgDataType)

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

func (s *PostgreSQLStorage) mapDataType(dataType string) string {
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

func (s *PostgreSQLStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
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

	s.mu.Lock()
	defer s.mu.Unlock()

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

	log.Printf("Executing query: %s with values: %v", query, values)
	_, err = s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *PostgreSQLStorage) buildUpdateClause(columns, placeholders []string) string {
	updateClauses := make([]string, 0, len(columns)-1)
	for i := 1; i < len(columns); i++ { // Skip id column
		updateClauses = append(updateClauses, fmt.Sprintf("%s = %s", columns[i], placeholders[i]))
	}
	return strings.Join(updateClauses, ", ")
}

func (s *PostgreSQLStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
	log.Printf("Executing query: %s with values: %v", query, values)

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

func (s *PostgreSQLStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.FindByKey(ctx, tblName, "id", id)
}

func (s *PostgreSQLStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	if tblName == "" || key == "" || value == "" {
		return nil, errors.New("table name, key, and value must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tbl, ok := s.sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	// Create ordered slices of fields to ensure consistent mapping
	fieldNames := make([]string, 0, len(tbl.Fields))
	for _, field := range tbl.Fields {
		fieldNames = append(fieldNames, field.Name)
	}

	columns := make([]string, 0, len(fieldNames))
	columnPointers := make([]any, 0, len(fieldNames))
	fieldTypes := make([]string, 0, len(fieldNames))

	for _, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		columns = append(columns, field.Name)
		fieldTypes = append(fieldTypes, field.DataType)

		log.Printf("Processing field: %s with data type: %s", field.Name, field.DataType)
		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, columnPointer)
			}
		case "INTEGER", "INT":
			if field.Nullable {
				columnPointer := new(*int64)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(int64)
				columnPointers = append(columnPointers, columnPointer)
			}
		case "REAL", "FLOAT":
			columnPointer := new(float64)
			columnPointers = append(columnPointers, columnPointer)
		case "BLOB":
			columnPointer := new([]byte)
			columnPointers = append(columnPointers, columnPointer)
		case "BOOLEAN":
			columnPointer := new(bool)
			columnPointers = append(columnPointers, columnPointer)
		case "JSON":
			columnPointer := new(string)
			columnPointers = append(columnPointers, columnPointer)
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			columnPointer := new(string)
			columnPointers = append(columnPointers, columnPointer)
		default:
			return nil, fmt.Errorf("unsupported data type %s for field %s", field.DataType, field.Name)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", strings.Join(columns, ", "), tbl.TableName, tbl.Fields[key].Name)

	log.Printf("Executing query: '%s' with value: %s", query, value)

	row := s.db.QueryRowContext(ctx, query, value)
	if err := row.Scan(columnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Create the object and populate fields after scanning
	var obj object.Object
	obj.TableName = tbl.TableName
	obj.Fields = make(map[string]any)

	for i, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		switch strings.ToUpper(fieldTypes[i]) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				if ptr, ok := columnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *columnPointers[i].(*string)
			}
		case "INTEGER", "INT":
			if field.Nullable {
				if ptr, ok := columnPointers[i].(**int64); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *columnPointers[i].(*int64)
			}
		case "REAL", "FLOAT":
			obj.Fields[field.Name] = *columnPointers[i].(*float64)
		case "BLOB":
			obj.Fields[field.Name] = *columnPointers[i].(*[]byte)
		case "BOOLEAN":
			obj.Fields[field.Name] = *columnPointers[i].(*bool)
		case "JSON":
			if field.Nullable {
				if ptr, ok := columnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *columnPointers[i].(*string)
			}
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			obj.Fields[field.Name] = *columnPointers[i].(*string)
		}
	}

	// Set the ID field from the object fields
	if idValue, exists := obj.Fields["id"]; exists && idValue != nil {
		obj.ID = idValue.(string)
	}

	return &obj, nil
}

func (s *PostgreSQLStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM `+tblName+` WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *PostgreSQLStorage) ResetConnection(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}