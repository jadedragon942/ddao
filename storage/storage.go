package storage

import (
	"context"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
)

type Storage interface {
	Connect(ctx context.Context, connStr string) error
	CreateTables(ctx context.Context, schema *schema.Schema) error
	Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error)
	Update(ctx context.Context, obj *object.Object) (bool, error)
	FindByID(ctx context.Context, tblName, id string) (*object.Object, error)
	FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error)
	DeleteByID(ctx context.Context, tblName, id string) (bool, error)
	ResetConnection(ctx context.Context) error
}
