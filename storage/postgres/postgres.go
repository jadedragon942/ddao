package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
	"github.com/jadedragon942/ddao/storage/common"
)

type PostgreSQLStorage struct {
	*common.BaseSQLStorage
}

func New() storage.Storage {
	return &PostgreSQLStorage{
		BaseSQLStorage: common.NewBaseSQLStorage(),
	}
}

func (s *PostgreSQLStorage) Connect(ctx context.Context, connStr string) error {
	// Parse connection string to create pgx config
	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		return err
	}

	// Create database/sql connection using pgx driver
	db := stdlib.OpenDB(*config)
	if err := db.PingContext(ctx); err != nil {
		return err
	}

	s.SetDB(db)
	return nil
}

func (s *PostgreSQLStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if err := s.ValidateConnection(); err != nil {
		return err
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

		storage.DebugLog(createTableQuery)
		log.Printf("Creating table %s with query: %s", table.TableName, createTableQuery)

		_, err := s.GetDB().ExecContext(ctx, createTableQuery)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}
	}

	s.SetSchema(schema)
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

	columns, placeholders, values, err := common.PrepareInsertData(obj, tbl, func(i int) string { return fmt.Sprintf("$%d", i) })
	if err != nil {
		return nil, false, err
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s",
		tbl.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		s.buildUpdateClause(columns, placeholders))

	storage.DebugLog(query, values...)
	_, err = s.GetDB().ExecContext(ctx, query, values...)
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
	if err := s.ValidateConnection(); err != nil {
		return false, err
	}

	tbl, err := s.GetTable(obj.TableName)
	if err != nil {
		return false, err
	}

	setClauses, values := common.PrepareUpdateData(obj, func(i int) string { return fmt.Sprintf("$%d", i) })

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tbl.TableName, strings.Join(setClauses, ", "), len(setClauses)+1)

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

func (s *PostgreSQLStorage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	// For PostgreSQL, Upsert is the same as Insert because Insert already uses ON CONFLICT DO UPDATE
	return s.Insert(ctx, obj)
}

func (s *PostgreSQLStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.BaseSQLStorage.FindByID(ctx, tblName, id, s.FindByKey)
}

func (s *PostgreSQLStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	return common.CommonFindByKey(ctx, s.GetDB(), s.GetSchema(), tblName, key, value, func(columns []string, tableName, keyField string) string {
		return fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", strings.Join(columns, ", "), tableName, keyField)
	})
}

func (s *PostgreSQLStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	return common.CommonDeleteByID(ctx, s.GetDB(), tblName, id, func(tableName string) string {
		return fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
	})
}

func (s *PostgreSQLStorage) ResetConnection(ctx context.Context) error {
	return s.BaseSQLStorage.ResetConnection(ctx)
}

// Transaction support methods
func (s *PostgreSQLStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.TransactionMethods.BeginTx(ctx)
}

func (s *PostgreSQLStorage) CommitTx(tx *sql.Tx) error {
	return s.TransactionMethods.CommitTx(tx)
}

func (s *PostgreSQLStorage) RollbackTx(tx *sql.Tx) error {
	return s.TransactionMethods.RollbackTx(tx)
}

func (s *PostgreSQLStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	if tx == nil {
		return nil, false, errors.New("transaction is nil")
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	if err := s.ValidateSchema(); err != nil {
		return nil, false, errors.New("schema not initialized")
	}

	tbl, ok := s.GetSchema().GetTable(obj.TableName)
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

func (s *PostgreSQLStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is nil")
	}

	tbl, ok := s.GetSchema().GetTable(obj.TableName)
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

func (s *PostgreSQLStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return s.FindByKeyTx(ctx, tx, tblName, "id", id)
}

func (s *PostgreSQLStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
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

func (s *PostgreSQLStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
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

func (s *PostgreSQLStorage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	// For PostgreSQL, UpsertTx is the same as InsertTx because InsertTx already uses ON CONFLICT DO UPDATE
	return s.InsertTx(ctx, tx, obj)
}
