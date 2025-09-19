package common

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
)

// ValidateConnection checks if the database connection is valid
func ValidateConnection(db *sql.DB) error {
	if db == nil {
		return errors.New("not connected")
	}
	return nil
}

// ValidateSchema checks if the schema is initialized
func ValidateSchema(sch *schema.Schema) error {
	if sch == nil {
		return errors.New("schema not initialized")
	}
	return nil
}

// ValidateFindByKeyParams validates parameters for FindByKey operations
func ValidateFindByKeyParams(tblName, key, value string) error {
	if tblName == "" || key == "" || value == "" {
		return errors.New("table name, key, and value must not be empty")
	}
	return nil
}

// ValidateTransaction checks if a transaction is valid
func ValidateTransaction(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return nil
}

// PrepareInsertData prepares columns, placeholders, and values for INSERT operations
func PrepareInsertData(obj *object.Object, tbl schema.TableSchema, placeholderFunc func(int) string) ([]string, []string, []any, error) {
	columns := make([]string, 0, len(obj.Fields)+1)
	placeholders := make([]string, 0, len(obj.Fields)+1)
	values := make([]any, 0, len(obj.Fields)+1)

	// Add ID field first
	columns = append(columns, "id")
	placeholders = append(placeholders, placeholderFunc(1))
	values = append(values, obj.ID)

	paramIndex := 2
	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue // Skip ID field, already handled
		}

		columns = append(columns, name)
		placeholders = append(placeholders, placeholderFunc(paramIndex))

		schField, ok := tbl.Fields[name]
		if !ok {
			return nil, nil, nil, fmt.Errorf("field %s not found in table %s schema", name, tbl.TableName)
		}

		if strings.ToLower(schField.DataType) == "json" {
			jsonData, err := json.Marshal(field)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to marshal JSON field %s: %w", name, err)
			}
			values = append(values, string(jsonData))
		} else {
			values = append(values, field)
		}
		paramIndex++
	}

	return columns, placeholders, values, nil
}

// PrepareUpdateData prepares SET clauses and values for UPDATE operations
func PrepareUpdateData(obj *object.Object, placeholderFunc func(int) string) ([]string, []any) {
	setClauses := make([]string, 0, len(obj.Fields)-1) // Exclude ID field
	values := make([]any, 0, len(obj.Fields))
	paramIndex := 1

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue // Skip ID field
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", name, placeholderFunc(paramIndex)))
		values = append(values, value)
		paramIndex++
	}

	// Add ID at the end for WHERE clause
	values = append(values, obj.ID)

	return setClauses, values
}

// CommonFindByKey implements common FindByKey logic for SQL databases
func CommonFindByKey(ctx context.Context, db *sql.DB, sch *schema.Schema, tblName, key, value string, queryFunc func([]string, string, string) string) (*object.Object, error) {
	if err := ValidateConnection(db); err != nil {
		return nil, err
	}

	if err := ValidateFindByKeyParams(tblName, key, value); err != nil {
		return nil, err
	}

	tbl, ok := sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	fieldScanner := NewFieldScanner(tbl)
	columns := fieldScanner.GetColumns()

	query := queryFunc(columns, tbl.TableName, tbl.Fields[key].Name)
	storage.DebugLog(query, value)

	row := db.QueryRowContext(ctx, query, value)
	if err := row.Scan(fieldScanner.ColumnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return fieldScanner.ScanToObject(tbl.TableName), nil
}

// CommonFindByKeyTx implements common FindByKey logic for SQL databases with transactions
func CommonFindByKeyTx(ctx context.Context, tx *sql.Tx, sch *schema.Schema, tblName, key, value string, queryFunc func([]string, string, string) string) (*object.Object, error) {
	if err := ValidateTransaction(tx); err != nil {
		return nil, err
	}

	if err := ValidateFindByKeyParams(tblName, key, value); err != nil {
		return nil, err
	}

	tbl, ok := sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	fieldScanner := NewFieldScanner(tbl)
	columns := fieldScanner.GetColumns()

	query := queryFunc(columns, tbl.TableName, tbl.Fields[key].Name)
	storage.DebugLog(query, value)

	row := tx.QueryRowContext(ctx, query, value)
	if err := row.Scan(fieldScanner.ColumnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return fieldScanner.ScanToObject(tbl.TableName), nil
}

// CommonDeleteByID implements common DeleteByID logic for SQL databases
func CommonDeleteByID(ctx context.Context, db *sql.DB, tblName, id string, queryFunc func(string) string, args ...any) (bool, error) {
	if err := ValidateConnection(db); err != nil {
		return false, err
	}

	query := queryFunc(tblName)
	allArgs := append([]any{id}, args...)
	storage.DebugLog(query, allArgs...)

	res, err := db.ExecContext(ctx, query, allArgs...)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// CommonDeleteByIDTx implements common DeleteByID logic for SQL databases with transactions
func CommonDeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string, queryFunc func(string) string, args ...any) (bool, error) {
	if err := ValidateTransaction(tx); err != nil {
		return false, err
	}

	query := queryFunc(tblName)
	allArgs := append([]any{id}, args...)
	storage.DebugLog(query, allArgs...)

	res, err := tx.ExecContext(ctx, query, allArgs...)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// TransactionMethods provides common transaction method implementations
type TransactionMethods struct {
	DB *sql.DB
}

// BeginTx starts a new transaction
func (tm *TransactionMethods) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if err := ValidateConnection(tm.DB); err != nil {
		return nil, err
	}
	return tm.DB.BeginTx(ctx, nil)
}

// CommitTx commits a transaction
func (tm *TransactionMethods) CommitTx(tx *sql.Tx) error {
	if err := ValidateTransaction(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// RollbackTx rolls back a transaction
func (tm *TransactionMethods) RollbackTx(tx *sql.Tx) error {
	if err := ValidateTransaction(tx); err != nil {
		return err
	}
	return tx.Rollback()
}