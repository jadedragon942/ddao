package orm

import (
	"testing"

	"github.com/jadedragon942/ddao/schema"
	sqliteStorage "github.com/jadedragon942/ddao/storage/sqlite"
)

func getTestSchema() *schema.Schema {
	sch := schema.New()

	table := schema.NewTableSchema("people")
	idField := schema.ColumnData{
		Name:          "id",
		DataType:      "integer",
		Nullable:      false,
		Default:       nil,
		Comment:       "Primary key for the person",
		Unique:        true,
		Index:         true,
		AutoIncrement: true,
	}
	table.AddField(idField)
	nameField := schema.ColumnData{
		Name:          "name",
		DataType:      "text",
		Nullable:      false,
		Default:       nil,
		Comment:       "Name of the person",
		Unique:        false,
		Index:         true,
		AutoIncrement: false,
	}
	table.AddField(nameField)

	metadataField := schema.ColumnData{
		Name:     "metadata",
		DataType: "json",
		Nullable: true,
		Default:  nil,
		Comment:  "Additional metadata for the person",
	}

	table.AddField(metadataField)

	sch.AddTable(table)

	return sch
}

func TestORM(t *testing.T) {
	sch := getTestSchema()
	o := New(sch)
	o.WithStorage(sqliteStorage.New())

}
