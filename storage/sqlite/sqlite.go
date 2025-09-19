package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
	"github.com/jadedragon942/ddao/storage/common"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	*common.BaseSQLStorage
}

func New() storage.Storage {
	return &SQLiteStorage{
		BaseSQLStorage: common.NewBaseSQLStorage(),
	}
}

func (s *SQLiteStorage) Connect(ctx context.Context, connStr string) error {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return err
	}
	s.SetDB(db)
	return nil
}

func (s *SQLiteStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if err := s.ValidateConnection(); err != nil {
		return err
	}

	for _, table := range schema.Tables {
		createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY", table.TableName)

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}
			createTableQuery += fmt.Sprintf(", %s %s", field.Name, field.DataType)
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

			/*
				if field.Index {
					createTableQuery += " INDEX"
				}
			*/
			if field.AutoIncrement {
				createTableQuery += " AUTOINCREMENT"
			}
			if field.PrimaryKey {
				createTableQuery += " PRIMARY KEY"
			}
		}

		createTableQuery += ")"

		storage.DebugLog(createTableQuery)

		_, err := s.GetDB().ExecContext(ctx, createTableQuery)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}
	}

	s.SetSchema(schema)

	return nil
}

func (s *SQLiteStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	if err := s.ValidateConnection(); err != nil {
		return nil, false, err
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	tbl, err := s.GetTable(obj.TableName)
	if err != nil {
		return nil, false, err
	}

	columns, placeholders, values, err := common.PrepareInsertData(obj, tbl, func(i int) string { return "?" })
	if err != nil {
		return nil, false, err
	}

	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		tbl.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	storage.DebugLog(query, values...)
	_, err = s.GetDB().ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *SQLiteStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
	if err := s.ValidateConnection(); err != nil {
		return false, err
	}

	tbl, err := s.GetTable(obj.TableName)
	if err != nil {
		return false, err
	}

	setClauses, values := common.PrepareUpdateData(obj, func(i int) string { return "?" })

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tbl.TableName, strings.Join(setClauses, ", "))
	storage.DebugLog(query, values...)

	res, err := s.GetDB().ExecContext(ctx, query, values...)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (s *SQLiteStorage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	// For SQLite, Upsert is the same as Insert because Insert already uses INSERT OR REPLACE
	return s.Insert(ctx, obj)
}

func (s *SQLiteStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.BaseSQLStorage.FindByID(ctx, tblName, id, s.FindByKey)
}

func (s *SQLiteStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	return common.CommonFindByKey(ctx, s.GetDB(), s.GetSchema(), tblName, key, value, func(columns []string, tableName, keyField string) string {
		return fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(columns, ", "), tableName, keyField)
	})
}

func (s *SQLiteStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	return common.CommonDeleteByID(ctx, s.GetDB(), tblName, id, func(tableName string) string {
		return fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	})
}

func (s *SQLiteStorage) ResetConnection(ctx context.Context) error {
	return s.BaseSQLStorage.ResetConnection(ctx)
}

// Transaction support methods
func (s *SQLiteStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.TransactionMethods.BeginTx(ctx)
}

func (s *SQLiteStorage) CommitTx(tx *sql.Tx) error {
	return s.TransactionMethods.CommitTx(tx)
}

func (s *SQLiteStorage) RollbackTx(tx *sql.Tx) error {
	return s.TransactionMethods.RollbackTx(tx)
}

func (s *SQLiteStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	if tx == nil {
		return nil, false, errors.New("transaction is nil")
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	// Ensure the table exists
	if err := s.ValidateSchema(); err != nil {
		return nil, false, errors.New("schema not initialized")
	}

	tbl, ok := s.GetSchema().GetTable(obj.TableName)
	if !ok {
		return nil, false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	bindingParams := ""
	columns := ""
	i := 0

	values := make([]any, 0, len(obj.Fields))
	columns += "id"
	bindingParams = "?"
	values = append(values, obj.ID)
	i++

	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			// Skip the ID field, it's handled separately
			continue
		}

		columns += fmt.Sprintf(", %s", name)
		bindingParams += ", ?"

		schField, ok := tbl.Fields[name]
		if !ok {
			return nil, false, fmt.Errorf("field %s not found in table %s schema", name, tbl.TableName)
		}
		if strings.ToLower(schField.DataType) == "json" {
			// If the field is JSON, we need to ensure it's stored as a string
			jsonData, err := json.Marshal(field)
			if err != nil {
				return nil, false, fmt.Errorf("failed to marshal JSON field %s: %w", name, err)
			}
			values = append(values, string(jsonData))
		} else {
			values = append(values, obj.Fields[schField.Name])
		}

		i++
	}

	query := `INSERT OR REPLACE INTO ` + tbl.TableName + ` (` + columns + `) VALUES (` + bindingParams + `)`
	storage.DebugLog(query, values...)
	_, err = tx.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *SQLiteStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}

	tbl, ok := s.GetSchema().GetTable(obj.TableName)
	if !ok {
		return false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	setClauses := make([]string, 0, len(obj.Fields)-1) // Exclude ID field
	values := make([]any, 0, len(obj.Fields))

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue // Skip the ID field
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", name))
		values = append(values, value)
	}

	values = append(values, obj.ID) // Add ID at the end for WHERE clause

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tbl.TableName, strings.Join(setClauses, ", "))
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

func (s *SQLiteStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return s.FindByKeyTx(ctx, tx, tblName, "id", id)
}

func (s *SQLiteStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tblName == "" || key == "" || value == "" {
		return nil, errors.New("table name, key, and value must not be empty")
	}

	tbl, ok := s.GetSchema().GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	// Use the field order from the table schema to ensure consistent mapping
	fieldNames := tbl.FieldOrder

	columns := make([]string, 0, len(fieldNames))
	columnPointers := make([]any, 0, len(fieldNames))
	fieldTypes := make([]string, 0, len(fieldNames))

	for _, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		columns = append(columns, field.Name)
		fieldTypes = append(fieldTypes, field.DataType)

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
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML", "JSON":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, columnPointer)
			}
		default:
			return nil, fmt.Errorf("unsupported data type %s for field %s", field.DataType, field.Name)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(columns, ", "), tbl.TableName, tbl.Fields[key].Name)

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
				// For nullable fields, use sql.NullString or check if pointer is nil
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
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML", "JSON":
			if field.Nullable {
				if ptr, ok := columnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *columnPointers[i].(*string)
			}
		}
	}

	// Set the ID field from the object fields
	if idValue, exists := obj.Fields["id"]; exists && idValue != nil {
		obj.ID = idValue.(string)
	}

	return &obj, nil
}

func (s *SQLiteStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}
	query := `DELETE FROM ` + tblName + ` WHERE id = ?`
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

func (s *SQLiteStorage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	// For SQLite, UpsertTx is the same as InsertTx because InsertTx already uses INSERT OR REPLACE
	return s.InsertTx(ctx, tx, obj)
}

func (s *SQLiteStorage) AlterTable(ctx context.Context, tableName, columnName, dataType string, nullable bool) error {
	if err := s.ValidateConnection(); err != nil {
		return err
	}

	nullableClause := "NOT NULL"
	if nullable {
		nullableClause = "NULL"
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s %s", tableName, columnName, dataType, nullableClause)

	storage.DebugLog(query)
	_, err := s.GetDB().ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to alter table %s: %w", tableName, err)
	}

	return nil
}
