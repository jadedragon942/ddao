package storage

import (
	"context"
	"database/sql"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
)

type Storage interface {
	Connect(ctx context.Context, connStr string) error
	CreateTables(ctx context.Context, schema *schema.Schema) error
	Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error)
	Update(ctx context.Context, obj *object.Object) (bool, error)
	Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error)
	FindByID(ctx context.Context, tblName, id string) (*object.Object, error)
	FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error)
	DeleteByID(ctx context.Context, tblName, id string) (bool, error)
	ResetConnection(ctx context.Context) error

	// Transaction support
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
	InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error)
	UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error)
	UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error)
	FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error)
	FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error)
	DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error)
}
