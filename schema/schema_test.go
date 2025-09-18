package schema

import (
	"testing"
)

func TestNewSchema(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected non-nil Schema")
	}
	if s.Tables == nil {
		t.Fatal("expected Tables map to be initialized")
	}
	if s.DatabaseName != "" {
		t.Errorf("expected empty DatabaseName, got %q", s.DatabaseName)
	}
}

func TestSetDatabaseName(t *testing.T) {
	s := New()
	s.SetDatabaseName("testdb")
	if s.DatabaseName != "testdb" {
		t.Errorf("expected DatabaseName to be 'testdb', got %q", s.DatabaseName)
	}
}

func TestGetTestSchema(t *testing.T) {
	s := GetTestSchema()
	if s == nil {
		t.Fatal("expected non-nil Schema from GetTestSchema")
	}
	if len(s.Tables) != 1 {
		t.Fatalf("expected 1 table in schema, got %d", len(s.Tables))
	}
}

func TestAddTableNil(t *testing.T) {
	s := New()
	s.AddTable(nil)
	if len(s.Tables) != 0 {
		t.Error("expected no tables to be added when nil is passed")
	}
}

func TestNewTableSchema(t *testing.T) {
	ts := NewTableSchema("products")
	if ts == nil {
		t.Fatal("expected non-nil TableSchema")
	}
	if ts.TableName != "products" {
		t.Errorf("expected TableName to be 'products', got %q", ts.TableName)
	}
	if ts.Fields == nil {
		t.Error("expected Fields map to be initialized")
	}
}

func TestAddField(t *testing.T) {
	ts := NewTableSchema("orders")
	field := ColumnData{
		Name:     "id",
		DataType: "int",
	}
	ts.AddField(field)
	if len(ts.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(ts.Fields))
	}
	if ts.Fields["id"].DataType != "int" {
		t.Errorf("expected DataType 'int', got %q", ts.Fields["id"].DataType)
	}
	if len(ts.FieldOrder) != 1 || ts.FieldOrder[0] != "id" {
		t.Errorf("expected FieldOrder to contain 'id'")
	}
}

func TestAddFieldEmptyName(t *testing.T) {
	ts := NewTableSchema("orders")
	field := ColumnData{
		Name:     "",
		DataType: "int",
	}
	ts.AddField(field)
	if len(ts.Fields) != 0 {
		t.Error("expected no fields to be added with empty name")
	}
}

func TestAddFieldPrimaryKeyAndAutoIncrement(t *testing.T) {
	ts := NewTableSchema("orders")
	field := ColumnData{
		Name:          "id",
		DataType:      "int",
		PrimaryKey:    true,
		AutoIncrement: true,
	}
	ts.AddField(field)
	if ts.PrimaryKey != "id" {
		t.Errorf("expected PrimaryKey to be 'id', got %q", ts.PrimaryKey)
	}
	if len(ts.AutoIncrementFields) != 1 || ts.AutoIncrementFields[0] != "id" {
		t.Errorf("expected AutoIncrementFields to contain 'id'")
	}
}

func TestMultipleFieldsOrder(t *testing.T) {
	ts := NewTableSchema("multi")
	fields := []ColumnData{
		{Name: "a", DataType: "int"},
		{Name: "b", DataType: "string"},
		{Name: "c", DataType: "bool"},
	}
	for _, f := range fields {
		ts.AddField(f)
	}
	if len(ts.FieldOrder) != 3 {
		t.Fatalf("expected 3 fields in FieldOrder, got %d", len(ts.FieldOrder))
	}
	for i, name := range []string{"a", "b", "c"} {
		if ts.FieldOrder[i] != name {
			t.Errorf("expected FieldOrder[%d] = %q, got %q", i, name, ts.FieldOrder[i])
		}
	}
}

func TestTableSchemaIndexesAndUniqueKeys(t *testing.T) {
	ts := NewTableSchema("test")
	ts.Indexes = []string{"idx1", "idx2"}
	ts.UniqueKeys = []string{"uk1"}
	if len(ts.Indexes) != 2 || ts.Indexes[0] != "idx1" || ts.Indexes[1] != "idx2" {
		t.Error("Indexes not set correctly")
	}
	if len(ts.UniqueKeys) != 1 || ts.UniqueKeys[0] != "uk1" {
		t.Error("UniqueKeys not set correctly")
	}
}

func TestColumnDataFields(t *testing.T) {
	col := ColumnData{
		Name:          "email",
		DataType:      "varchar",
		Nullable:      true,
		Default:       "none",
		Comment:       "user email",
		Unique:        true,
		Index:         true,
		AutoIncrement: false,
		PrimaryKey:    false,
	}
	if col.Name != "email" {
		t.Errorf("expected Name to be 'email', got %q", col.Name)
	}
	if col.DataType != "varchar" {
		t.Errorf("expected DataType to be 'varchar', got %q", col.DataType)
	}
	if !col.Nullable {
		t.Error("expected Nullable to be true")
	}
	if col.Default != "none" {
		t.Errorf("expected Default to be 'none', got %v", col.Default)
	}
	if !col.Unique {
		t.Error("expected Unique to be true")
	}
	if !col.Index {
		t.Error("expected Index to be true")
	}
	if col.AutoIncrement {
		t.Error("expected AutoIncrement to be false")
	}
	if col.PrimaryKey {
		t.Error("expected PrimaryKey to be false")
	}
	if col.Comment != "user email" {
		t.Errorf("expected Comment to be 'user email', got %q", col.Comment)
	}
}
