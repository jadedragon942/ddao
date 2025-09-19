package common

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
)

// BaseSQLStorage provides common functionality for SQL-based storage adapters
type BaseSQLStorage struct {
	DB             *sql.DB
	Sch            *schema.Schema
	TransactionMethods
}

// NewBaseSQLStorage creates a new base SQL storage instance
func NewBaseSQLStorage() *BaseSQLStorage {
	base := &BaseSQLStorage{}
	base.TransactionMethods.DB = base.DB
	return base
}

// SetDB sets the database connection and updates transaction methods
func (b *BaseSQLStorage) SetDB(db *sql.DB) {
	b.DB = db
	b.TransactionMethods.DB = db
}

// SetSchema sets the schema
func (b *BaseSQLStorage) SetSchema(sch *schema.Schema) {
	b.Sch = sch
}

// GetSchema returns the current schema
func (b *BaseSQLStorage) GetSchema() *schema.Schema {
	return b.Sch
}

// GetDB returns the database connection
func (b *BaseSQLStorage) GetDB() *sql.DB {
	return b.DB
}

// ResetConnection closes the database connection
func (b *BaseSQLStorage) ResetConnection(ctx context.Context) error {
	if b.DB != nil {
		return b.DB.Close()
	}
	return nil
}

// FindByID is a common implementation that delegates to FindByKey
func (b *BaseSQLStorage) FindByID(ctx context.Context, tblName, id string, findByKeyFunc func(context.Context, string, string, string) (*object.Object, error)) (*object.Object, error) {
	return findByKeyFunc(ctx, tblName, "id", id)
}

// FindByIDTx is a common implementation that delegates to FindByKeyTx
func (b *BaseSQLStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string, findByKeyTxFunc func(context.Context, *sql.Tx, string, string, string) (*object.Object, error)) (*object.Object, error) {
	return findByKeyTxFunc(ctx, tx, tblName, "id", id)
}

// ValidateConnection validates the database connection
func (b *BaseSQLStorage) ValidateConnection() error {
	return ValidateConnection(b.DB)
}

// ValidateSchema validates the schema
func (b *BaseSQLStorage) ValidateSchema() error {
	return ValidateSchema(b.Sch)
}

// GetTable safely gets a table from the schema
func (b *BaseSQLStorage) GetTable(tableName string) (schema.TableSchema, error) {
	if err := b.ValidateSchema(); err != nil {
		return schema.TableSchema{}, err
	}

	tbl, ok := b.Sch.GetTable(tableName)
	if !ok {
		return schema.TableSchema{}, errors.New("table " + tableName + " not found in schema")
	}

	return tbl, nil
}