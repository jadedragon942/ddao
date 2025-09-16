package orm

import (
	"context"

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
