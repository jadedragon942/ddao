package orm

import (
	"context"
	"database/sql"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
)

type ORM struct {
	Schema  *schema.Schema
	Storage storage.Storage
}

func New(schema *schema.Schema) *ORM {
	return &ORM{
		Schema: schema,
	}
}

func (orm *ORM) WithStorage(storage storage.Storage) *ORM {
	orm.Storage = storage
	return orm
}

func (orm *ORM) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return orm.Storage.Insert(ctx, obj)
}

func (orm *ORM) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return orm.Storage.Upsert(ctx, obj)
}

func (orm *ORM) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return orm.Storage.FindByKey(ctx, tblName, "id", id)
}

func (orm *ORM) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	return orm.Storage.FindByKey(ctx, tblName, key, value)
}

func (orm *ORM) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	return orm.Storage.DeleteByID(ctx, tblName, id)
}

func (orm *ORM) ResetConnection(ctx context.Context) error {
	return orm.Storage.ResetConnection(ctx)
}

func (orm *ORM) Connect(ctx context.Context, connStr string) error {
	return orm.Storage.Connect(ctx, connStr)
}

// Transaction support methods
func (orm *ORM) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return orm.Storage.BeginTx(ctx)
}

func (orm *ORM) CommitTx(tx *sql.Tx) error {
	return orm.Storage.CommitTx(tx)
}

func (orm *ORM) RollbackTx(tx *sql.Tx) error {
	return orm.Storage.RollbackTx(tx)
}

func (orm *ORM) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return orm.Storage.InsertTx(ctx, tx, obj)
}

func (orm *ORM) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	return orm.Storage.UpdateTx(ctx, tx, obj)
}

func (orm *ORM) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return orm.Storage.UpsertTx(ctx, tx, obj)
}

func (orm *ORM) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return orm.Storage.FindByIDTx(ctx, tx, tblName, id)
}

func (orm *ORM) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	return orm.Storage.FindByKeyTx(ctx, tx, tblName, key, value)
}

func (orm *ORM) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	return orm.Storage.DeleteByIDTx(ctx, tx, tblName, id)
}
