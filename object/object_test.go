package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	obj := New()
	assert.NotNil(t, obj)
	assert.Empty(t, obj.TableName)
	assert.NotNil(t, obj.Fields)
	assert.Empty(t, obj.Fields)
}

func TestSetGetField(t *testing.T) {
	obj := New()
	obj.SetField("name", "test")
	value, exists := obj.GetField("name")
	assert.True(t, exists)
	assert.Equal(t, "test", value)

	// Test non-existing field
	_, exists = obj.GetField("non_existing")
	assert.False(t, exists)
}

func TestSetGetTableName(t *testing.T) {
	obj := New()
	obj.SetTableName("test_table")
	assert.Equal(t, "test_table", obj.GetTableName())

	// Test empty table name
	obj.SetTableName("")
	assert.Equal(t, "", obj.GetTableName())
}

func TestSetGetFields(t *testing.T) {
	obj := New()
	fields := map[string]interface{}{
		"name": "test",
		"age":  30,
	}
	obj.SetFields(fields)
	assert.Equal(t, fields, obj.GetFields())

	// Test empty fields
	obj.SetFields(nil)
	assert.Nil(t, obj.GetFields())
}

func TestGetString(t *testing.T) {
	obj := New()
	obj.SetField("name", "test")
	value, exists := obj.GetString("name")
	assert.True(t, exists)
	assert.Equal(t, "test", value)

	// Test non-string field
	obj.SetField("age", 30)
	_, exists = obj.GetString("age")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetString("non_existing")
	assert.False(t, exists)
}

func TestGetInt64(t *testing.T) {
	obj := New()
	obj.SetField("age", int64(30))
	value, exists := obj.GetInt64("age")
	assert.True(t, exists)
	assert.Equal(t, int64(30), value)

	// Test non-int64 field
	obj.SetField("name", "test")
	_, exists = obj.GetInt64("name")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetInt64("non_existing")
	assert.False(t, exists)
}

func TestGetInt(t *testing.T) {
	obj := New()
	obj.SetField("age", 30)
	value, exists := obj.GetInt("age")
	assert.True(t, exists)
	assert.Equal(t, 30, value)

	// Test non-int field
	obj.SetField("name", "test")
	_, exists = obj.GetInt("name")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetInt("non_existing")
	assert.False(t, exists)
}

func TestGetUint64(t *testing.T) {
	obj := New()
	obj.SetField("count", uint64(30))
	value, exists := obj.GetUint64("count")
	assert.True(t, exists)
	assert.Equal(t, uint64(30), value)

	// Test non-uint64 field
	obj.SetField("name", "test")
	_, exists = obj.GetUint64("name")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetUint64("non_existing")
	assert.False(t, exists)
}

func TestGetFloat64(t *testing.T) {
	obj := New()
	obj.SetField("price", 19.99)
	value, exists := obj.GetFloat64("price")
	assert.True(t, exists)
	assert.Equal(t, 19.99, value)

	// Test non-float64 field
	obj.SetField("name", "test")
	_, exists = obj.GetFloat64("name")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetFloat64("non_existing")
	assert.False(t, exists)
}

func TestGetBool(t *testing.T) {
	obj := New()
	obj.SetField("active", true)
	value, exists := obj.GetBool("active")
	assert.True(t, exists)
	assert.Equal(t, true, value)

	// Test non-bool field
	obj.SetField("name", "test")
	_, exists = obj.GetBool("name")
	assert.False(t, exists)

	// Test non-existing field
	_, exists = obj.GetBool("non_existing")
	assert.False(t, exists)
}
func TestSetField(t *testing.T) {
	obj := New()
	obj.SetField("name", "test")
	value, exists := obj.GetField("name")
	assert.True(t, exists)
	assert.Equal(t, "test", value)

	// Overwrite existing field
	obj.SetField("name", "updated_test")
	value, exists = obj.GetField("name")
	assert.True(t, exists)
	assert.Equal(t, "updated_test", value)

	// Set a new field
	obj.SetField("age", 30)
	value, exists = obj.GetField("age")
	assert.True(t, exists)
	assert.Equal(t, 30, value)
}
func TestSetFields(t *testing.T) {
	obj := New()
	fields := map[string]interface{}{
		"name": "test",
		"age":  30,
	}
	obj.SetFields(fields)
	assert.Equal(t, fields, obj.GetFields())

	// Overwrite existing fields
	newFields := map[string]interface{}{
		"name": "updated_test",
		"city": "New York",
	}
	obj.SetFields(newFields)
	assert.Equal(t, newFields, obj.GetFields())

	// Set empty fields
	obj.SetFields(nil)
	assert.Nil(t, obj.GetFields())
}
func TestGetField(t *testing.T) {
	obj := New()
	obj.SetField("name", "test")
	value, exists := obj.GetField("name")
	assert.True(t, exists)
	assert.Equal(t, "test", value)

	// Test non-existing field
	_, exists = obj.GetField("non_existing")
	assert.False(t, exists)

	// Test empty fields
	obj.SetFields(nil)
	_, exists = obj.GetField("name")
	assert.False(t, exists)
}

func TestGetTableName(t *testing.T) {
	obj := New()
	obj.SetTableName("test_table")
	assert.Equal(t, "test_table", obj.GetTableName())

	// Test empty table name
	obj.SetTableName("")
	assert.Equal(t, "", obj.GetTableName())
}
