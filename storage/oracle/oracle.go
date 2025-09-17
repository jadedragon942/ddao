package oracle

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
	_ "github.com/godror/godror"
)

type OracleStorage struct {
	db  *sql.DB
	sch *schema.Schema
	mu  *sync.Mutex
}

func New() storage.Storage {
	var mu sync.Mutex
	return &OracleStorage{mu: &mu}
}

func (s *OracleStorage) Connect(ctx context.Context, connStr string) error {
	db, err := sql.Open("godror", connStr)
	if err != nil {
		return err
	}

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *OracleStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.db == nil {
		return errors.New("not connected")
	}

	for _, table := range schema.Tables {
		// Check if table exists
		var count int
		checkQuery := "SELECT COUNT(*) FROM user_tables WHERE table_name = UPPER(?)"
		err := s.db.QueryRowContext(ctx, checkQuery, table.TableName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check if table exists: %w", err)
		}

		if count > 0 {
			log.Printf("Table %s already exists, skipping creation", table.TableName)
			continue
		}

		createTableQuery := fmt.Sprintf("CREATE TABLE %s (id VARCHAR2(255) PRIMARY KEY", strings.ToUpper(table.TableName))

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}

			// Map data types to Oracle equivalents
			oracleDataType := s.mapDataType(field.DataType)
			createTableQuery += fmt.Sprintf(", %s %s", strings.ToUpper(field.Name), oracleDataType)

			if !field.Nullable {
				createTableQuery += " NOT NULL"
			}
			if field.Default != nil {
				defaultValue := field.Default
				if strings.ToUpper(field.DataType) == "TEXT" || strings.ToUpper(field.DataType) == "VARCHAR" {
					defaultValue = fmt.Sprintf("'%v'", defaultValue)
				}
				createTableQuery += fmt.Sprintf(" DEFAULT %v", defaultValue)
			}
		}

		createTableQuery += ")"

		log.Printf("Creating table %s with query: %s", table.TableName, createTableQuery)

		_, err = s.db.ExecContext(ctx, createTableQuery)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}

		// Create unique constraints separately for fields marked as unique
		for _, field := range table.Fields {
			if field.Unique && field.Name != "id" {
				constraintName := fmt.Sprintf("UK_%s_%s", strings.ToUpper(table.TableName), strings.ToUpper(field.Name))
				uniqueQuery := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
					strings.ToUpper(table.TableName), constraintName, strings.ToUpper(field.Name))
				_, err = s.db.ExecContext(ctx, uniqueQuery)
				if err != nil {
					log.Printf("Warning: failed to create unique constraint for %s.%s: %v", table.TableName, field.Name, err)
				}
			}
		}
	}

	s.sch = schema
	log.Println("Tables created successfully")

	return nil
}

func (s *OracleStorage) mapDataType(dataType string) string {
	switch strings.ToUpper(dataType) {
	case "TEXT", "VARCHAR", "CHAR":
		return "CLOB"
	case "INTEGER", "INT":
		return "NUMBER(19)"
	case "REAL", "FLOAT":
		return "BINARY_DOUBLE"
	case "BLOB":
		return "BLOB"
	case "BOOLEAN":
		return "NUMBER(1)" // Oracle doesn't have native boolean, use 0/1
	case "JSON":
		return "CLOB" // Oracle 12c+ has JSON support but we'll use CLOB for compatibility
	case "DATETIME", "TIMESTAMP":
		return "TIMESTAMP"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIMESTAMP"
	case "UUID":
		return "VARCHAR2(36)"
	default:
		return "CLOB"
	}
}

func (s *OracleStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
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
	updateClauses := make([]string, 0, len(obj.Fields))

	columns = append(columns, "ID")
	placeholders = append(placeholders, ":1")
	values = append(values, obj.ID)

	paramIndex := 2
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, strings.ToUpper(name))
		placeholders = append(placeholders, fmt.Sprintf(":%d", paramIndex))
		updateClauses = append(updateClauses, fmt.Sprintf("%s = :%d", strings.ToUpper(name), paramIndex+len(obj.Fields)))

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
		} else if strings.ToLower(schField.DataType) == "boolean" {
			// Convert boolean to number for Oracle
			if boolVal, ok := field.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, field)
			}
		} else {
			values = append(values, field)
		}
		paramIndex++
	}

	// Add values again for the UPDATE part of MERGE
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
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
		} else if strings.ToLower(schField.DataType) == "boolean" {
			// Convert boolean to number for Oracle
			if boolVal, ok := field.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, field)
			}
		} else {
			values = append(values, field)
		}
	}

	// Use MERGE for UPSERT functionality
	query := fmt.Sprintf(`
		MERGE INTO %s target
		USING (SELECT %s FROM dual) source
		ON (target.ID = source.ID)
		WHEN MATCHED THEN
			UPDATE SET %s
		WHEN NOT MATCHED THEN
			INSERT (%s) VALUES (%s)`,
		strings.ToUpper(tbl.TableName),
		func() string {
			selectParts := make([]string, 0, len(columns))
			for i, col := range columns {
				selectParts = append(selectParts, fmt.Sprintf("%s AS %s", placeholders[i], col))
			}
			return strings.Join(selectParts, ", ")
		}(),
		strings.Join(updateClauses, ", "),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	log.Printf("Executing query: %s with values: %v", query, values)
	_, err = s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *OracleStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
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
		setClauses = append(setClauses, fmt.Sprintf("%s = :%d", strings.ToUpper(name), paramIndex))

		// Handle boolean conversion for Oracle
		schField, ok := tbl.Fields[name]
		if ok && strings.ToLower(schField.DataType) == "boolean" {
			if boolVal, ok := value.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, value)
			}
		} else {
			values = append(values, value)
		}
		paramIndex++
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = :%d", strings.ToUpper(tbl.TableName), strings.Join(setClauses, ", "), paramIndex)
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

func (s *OracleStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.FindByKey(ctx, tblName, "id", id)
}

func (s *OracleStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
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
		columns = append(columns, strings.ToUpper(field.Name))
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
			columnPointer := new(int64) // Oracle stores boolean as number
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

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = :1", strings.Join(columns, ", "), strings.ToUpper(tbl.TableName), strings.ToUpper(key))

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
			// Convert Oracle number back to boolean
			numVal := *columnPointers[i].(*int64)
			obj.Fields[field.Name] = numVal != 0
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

func (s *OracleStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE ID = :1", strings.ToUpper(tblName))
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

func (s *OracleStorage) ResetConnection(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Transaction support methods
func (s *OracleStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, errors.New("not connected")
	}
	return s.db.BeginTx(ctx, nil)
}

func (s *OracleStorage) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Commit()
}

func (s *OracleStorage) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Rollback()
}

func (s *OracleStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
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

	s.mu.Lock()
	defer s.mu.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return nil, false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	columns := make([]string, 0, len(obj.Fields)+1)
	placeholders := make([]string, 0, len(obj.Fields)+1)
	values := make([]any, 0, len(obj.Fields)+1)
	updateClauses := make([]string, 0, len(obj.Fields))

	columns = append(columns, "ID")
	placeholders = append(placeholders, ":1")
	values = append(values, obj.ID)

	paramIndex := 2
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, strings.ToUpper(name))
		placeholders = append(placeholders, fmt.Sprintf(":%d", paramIndex))
		updateClauses = append(updateClauses, fmt.Sprintf("%s = :%d", strings.ToUpper(name), paramIndex+len(obj.Fields)))

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
		} else if strings.ToLower(schField.DataType) == "boolean" {
			// Convert boolean to number for Oracle
			if boolVal, ok := field.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, field)
			}
		} else {
			values = append(values, field)
		}
		paramIndex++
	}

	// Add values again for the UPDATE part of MERGE
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
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
		} else if strings.ToLower(schField.DataType) == "boolean" {
			// Convert boolean to number for Oracle
			if boolVal, ok := field.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, field)
			}
		} else {
			values = append(values, field)
		}
	}

	// Use MERGE for UPSERT functionality
	query := fmt.Sprintf(`
		MERGE INTO %s target
		USING (SELECT %s FROM dual) source
		ON (target.ID = source.ID)
		WHEN MATCHED THEN
			UPDATE SET %s
		WHEN NOT MATCHED THEN
			INSERT (%s) VALUES (%s)`,
		strings.ToUpper(tbl.TableName),
		func() string {
			selectParts := make([]string, 0, len(columns))
			for i, col := range columns {
				selectParts = append(selectParts, fmt.Sprintf("%s AS %s", placeholders[i], col))
			}
			return strings.Join(selectParts, ", ")
		}(),
		strings.Join(updateClauses, ", "),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	log.Printf("Executing transaction query: %s with values: %v", query, values)
	_, err = tx.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *OracleStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
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
		setClauses = append(setClauses, fmt.Sprintf("%s = :%d", strings.ToUpper(name), paramIndex))

		// Handle boolean conversion for Oracle
		schField, ok := tbl.Fields[name]
		if ok && strings.ToLower(schField.DataType) == "boolean" {
			if boolVal, ok := value.(bool); ok {
				if boolVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, value)
			}
		} else {
			values = append(values, value)
		}
		paramIndex++
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = :%d", strings.ToUpper(tbl.TableName), strings.Join(setClauses, ", "), paramIndex)
	log.Printf("Executing transaction query: %s with values: %v", query, values)

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

func (s *OracleStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return s.FindByKeyTx(ctx, tx, tblName, "id", id)
}

func (s *OracleStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
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
		columns = append(columns, strings.ToUpper(field.Name))
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
			columnPointer := new(int64) // Oracle stores boolean as number
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

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = :1", strings.Join(columns, ", "), strings.ToUpper(tbl.TableName), strings.ToUpper(key))

	log.Printf("Executing transaction query: '%s' with value: %s", query, value)

	row := tx.QueryRowContext(ctx, query, value)
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
			// Convert Oracle number back to boolean
			numVal := *columnPointers[i].(*int64)
			obj.Fields[field.Name] = numVal != 0
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

func (s *OracleStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE ID = :1", strings.ToUpper(tblName))
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