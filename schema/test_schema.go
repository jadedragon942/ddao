package schema

func GetTestSchema() *Schema {
	sch := New()

	table := NewTableSchema("people")
	idField := ColumnData{
		Name:          "id",
		DataType:      "text",
		Nullable:      false,
		Default:       nil,
		Comment:       "Primary key for the person",
		Unique:        true,
		Index:         true,
		AutoIncrement: true,
	}
	table.AddField(idField)
	nameField := ColumnData{
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

	metadataField := ColumnData{
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
