package sqlserver

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
	_ "github.com/microsoft/go-mssqldb"
)

type SQLServerStorage struct {
	db  *sql.DB
	sch *schema.Schema
	mu  *sync.Mutex
}

func New() storage.Storage {
	var mu sync.Mutex
	return &SQLServerStorage{mu: &mu}
}

func (s *SQLServerStorage) Connect(ctx context.Context, connStr string) error {
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return err
	}

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *SQLServerStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.db == nil {
		return errors.New("not connected")
	}

	for _, table := range schema.Tables {
		createTableQuery := fmt.Sprintf("IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='%s' AND xtype='U') CREATE TABLE [%s] ([id] NVARCHAR(255) PRIMARY KEY", table.TableName, table.TableName)

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}

			// Map data types to SQL Server equivalents
			sqlServerDataType := s.mapDataType(field.DataType)
			createTableQuery += fmt.Sprintf(", [%s] %s", field.Name, sqlServerDataType)

			if field.Nullable {
				createTableQuery += " NULL"
			} else {
				createTableQuery += " NOT NULL"
			}
			if field.Default != nil {
				defaultValue := field.Default
				if strings.ToUpper(field.DataType) == "TEXT" || strings.ToUpper(field.DataType) == "VARCHAR" {
					defaultValue = fmt.Sprintf("'%v'", defaultValue)
				}
				createTableQuery += fmt.Sprintf(" DEFAULT %v", defaultValue)
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

func (s *SQLServerStorage) mapDataType(dataType string) string {
	switch strings.ToUpper(dataType) {
	case "TEXT", "VARCHAR", "CHAR":
		return "NVARCHAR(MAX)"
	case "INTEGER", "INT":
		return "BIGINT"
	case "REAL", "FLOAT":
		return "FLOAT"
	case "BLOB":
		return "VARBINARY(MAX)"
	case "BOOLEAN":
		return "BIT"
	case "JSON":
		return "NVARCHAR(MAX)" // SQL Server 2016+ has JSON support but we'll use NVARCHAR for compatibility
	case "DATETIME", "TIMESTAMP":
		return "DATETIME2"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIME"
	case "UUID":
		return "UNIQUEIDENTIFIER"
	default:
		return "NVARCHAR(MAX)"
	}
}

func (s *SQLServerStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
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

	columns = append(columns, "[id]")
	placeholders = append(placeholders, "?")
	values = append(values, obj.ID)

	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, fmt.Sprintf("[%s]", name))
		placeholders = append(placeholders, "?")
		updateClauses = append(updateClauses, fmt.Sprintf("[%s] = ?", name))

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
		} else {
			values = append(values, field)
		}
	}

	// Use MERGE for UPSERT functionality
	query := fmt.Sprintf(`
		MERGE [%s] AS target
		USING (SELECT %s AS [id] %s) AS source
		ON target.[id] = source.[id]
		WHEN MATCHED THEN
			UPDATE SET %s
		WHEN NOT MATCHED THEN
			INSERT (%s) VALUES (%s);`,
		tbl.TableName,
		"?",
		func() string {
			if len(columns) > 1 {
				selectParts := make([]string, 0, len(columns)-1)
				for i := 1; i < len(columns); i++ {
					selectParts = append(selectParts, fmt.Sprintf(", ? AS %s", columns[i]))
				}
				return strings.Join(selectParts, "")
			}
			return ""
		}(),
		strings.Join(updateClauses, ", "),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	storage.DebugLog(query, values...)
	_, err = s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *SQLServerStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
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

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("[%s] = ?", name))
		values = append(values, value)
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE [%s] SET %s WHERE [id] = ?", tbl.TableName, strings.Join(setClauses, ", "))
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

// Upsert inserts or updates an object, delegating to Insert which already implements upsert behavior using MERGE statement
func (s *SQLServerStorage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return s.Insert(ctx, obj)
}

// UpsertTx inserts or updates an object within a transaction, delegating to InsertTx which already implements upsert behavior using MERGE statement
func (s *SQLServerStorage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return s.InsertTx(ctx, tx, obj)
}

func (s *SQLServerStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.FindByKey(ctx, tblName, "id", id)
}

func (s *SQLServerStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
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
		columns = append(columns, fmt.Sprintf("[%s]", field.Name))
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

	query := fmt.Sprintf("SELECT %s FROM [%s] WHERE [%s] = ?", strings.Join(columns, ", "), tbl.TableName, key)

	storage.DebugLog(query, value)

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

func (s *SQLServerStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("not connected")
	}
	query := fmt.Sprintf(`DELETE FROM [%s] WHERE [id] = ?`, tblName)
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

func (s *SQLServerStorage) ResetConnection(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Transaction support methods
func (s *SQLServerStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, errors.New("not connected")
	}
	return s.db.BeginTx(ctx, nil)
}

func (s *SQLServerStorage) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Commit()
}

func (s *SQLServerStorage) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return tx.Rollback()
}

func (s *SQLServerStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
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

	columns = append(columns, "[id]")
	placeholders = append(placeholders, "?")
	values = append(values, obj.ID)

	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, fmt.Sprintf("[%s]", name))
		placeholders = append(placeholders, "?")
		updateClauses = append(updateClauses, fmt.Sprintf("[%s] = ?", name))

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
		} else {
			values = append(values, field)
		}
	}

	// Use MERGE for UPSERT functionality
	query := fmt.Sprintf(`
		MERGE [%s] AS target
		USING (SELECT %s AS [id] %s) AS source
		ON target.[id] = source.[id]
		WHEN MATCHED THEN
			UPDATE SET %s
		WHEN NOT MATCHED THEN
			INSERT (%s) VALUES (%s);`,
		tbl.TableName,
		"?",
		func() string {
			if len(columns) > 1 {
				selectParts := make([]string, 0, len(columns)-1)
				for i := 1; i < len(columns); i++ {
					selectParts = append(selectParts, fmt.Sprintf(", ? AS %s", columns[i]))
				}
				return strings.Join(selectParts, "")
			}
			return ""
		}(),
		strings.Join(updateClauses, ", "),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	storage.DebugLog(query, values...)
	_, err = tx.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *SQLServerStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
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

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("[%s] = ?", name))
		values = append(values, value)
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE [%s] SET %s WHERE [id] = ?", tbl.TableName, strings.Join(setClauses, ", "))
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

func (s *SQLServerStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return s.FindByKeyTx(ctx, tx, tblName, "id", id)
}

func (s *SQLServerStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
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
		columns = append(columns, fmt.Sprintf("[%s]", field.Name))
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

	query := fmt.Sprintf("SELECT %s FROM [%s] WHERE [%s] = ?", strings.Join(columns, ", "), tbl.TableName, key)

	storage.DebugLog(query, value)

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

func (s *SQLServerStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}
	query := fmt.Sprintf(`DELETE FROM [%s] WHERE [id] = ?`, tblName)
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
