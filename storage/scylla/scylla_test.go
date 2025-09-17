package scylla

import (
	"context"
	"testing"
	"time"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("test_db")

	// Create test table
	testTable := schema.NewTableSchema("users")

	// Add fields
	testTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "User ID",
	})

	testTable.AddField(schema.ColumnData{
		Name:     "email",
		DataType: "text",
		Nullable: false,
		Comment:  "User email address",
	})

	testTable.AddField(schema.ColumnData{
		Name:     "age",
		DataType: "integer",
		Nullable: true,
		Comment:  "User age",
	})

	testTable.AddField(schema.ColumnData{
		Name:     "active",
		DataType: "boolean",
		Nullable: false,
		Comment:  "User active status",
	})

	testTable.AddField(schema.ColumnData{
		Name:     "profile",
		DataType: "json",
		Nullable: true,
		Comment:  "User profile data",
	})

	testTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	sch.AddTable(testTable)
	return sch
}

func TestScyllaDBStorageBasic(t *testing.T) {
	storage := New()
	assert.NotNil(t, storage)

	scyllaStorage, ok := storage.(*ScyllaDBStorage)
	assert.True(t, ok)
	assert.NotNil(t, scyllaStorage.mu)
}

func TestScyllaDBParseConsistency(t *testing.T) {
	storage := &ScyllaDBStorage{}

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"one", "ONE", false},
		{"quorum", "QUORUM", false},
		{"all", "ALL", false},
		{"localquorum", "LOCAL_QUORUM", false},
		{"invalid", "", true},
	}

	for _, test := range tests {
		consistency, err := storage.parseConsistency(test.input)
		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, consistency.String())
		}
	}
}

func TestScyllaDBMapDataType(t *testing.T) {
	storage := &ScyllaDBStorage{}

	tests := []struct {
		input    string
		expected string
	}{
		{"text", "text"},
		{"TEXT", "text"},
		{"integer", "bigint"},
		{"INT", "bigint"},
		{"boolean", "boolean"},
		{"json", "text"},
		{"datetime", "timestamp"},
		{"uuid", "uuid"},
		{"unknown", "text"},
	}

	for _, test := range tests {
		result := storage.mapDataType(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestScyllaDBConnect_InvalidConnectionString(t *testing.T) {
	storage := New()
	ctx := context.Background()

	// Test invalid connection string format
	err := storage.Connect(ctx, "invalid-connection-string")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid connection string format")
}

func TestScyllaDBOperations_NotConnected(t *testing.T) {
	storage := New()
	ctx := context.Background()
	schema := createTestSchema()

	// Test CreateTables without connection
	err := storage.CreateTables(ctx, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// Test Insert without connection
	obj := object.New()
	obj.TableName = "users"
	obj.ID = "test-id"
	_, _, err = storage.Insert(ctx, obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// Test Update without connection
	_, err = storage.Update(ctx, obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// Test FindByID without connection
	_, err = storage.FindByID(ctx, "users", "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// Test DeleteByID without connection
	_, err = storage.DeleteByID(ctx, "users", "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestScyllaDBTransactionMethods(t *testing.T) {
	storage := New()
	ctx := context.Background()

	// Test that all transaction methods return appropriate errors
	tx, err := storage.BeginTx(ctx)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	err = storage.CommitTx(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	err = storage.RollbackTx(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	obj := object.New()
	_, _, err = storage.InsertTx(ctx, nil, obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	_, err = storage.UpdateTx(ctx, nil, obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	_, err = storage.FindByIDTx(ctx, nil, "users", "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	_, err = storage.FindByKeyTx(ctx, nil, "users", "id", "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")

	_, err = storage.DeleteByIDTx(ctx, nil, "users", "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support SQL-style transactions")
}

// TestScyllaDBLocal tests against a local ScyllaDB instance
// This test is skipped unless explicitly run with a tag
func TestScyllaDBLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ScyllaDB integration test in short mode")
	}

	storage := New()
	ctx := context.Background()

	// Try to connect to local ScyllaDB instance
	err := storage.Connect(ctx, "localhost:9042/testks?consistency=one&timeout=5s")
	if err != nil {
		t.Skipf("ScyllaDB not available: %v", err)
	}
	defer storage.ResetConnection(ctx)

	schema := createTestSchema()

	// Test CreateTables
	err = storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Test Insert
	obj := object.New()
	obj.TableName = "users"
	obj.ID = "test-user-1"
	obj.Fields = map[string]interface{}{
		"email":      "test@example.com",
		"age":        int64(30),
		"active":     true,
		"profile":    `{"name": "Test User", "location": "Earth"}`,
		"created_at": time.Now(),
	}

	data, created, err := storage.Insert(ctx, obj)
	require.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, data)

	// Test FindByID
	foundObj, err := storage.FindByID(ctx, "users", "test-user-1")
	require.NoError(t, err)
	require.NotNil(t, foundObj)
	assert.Equal(t, "test-user-1", foundObj.ID)
	assert.Equal(t, "test@example.com", foundObj.Fields["email"])

	// Test Update
	foundObj.Fields["email"] = "updated@example.com"
	foundObj.Fields["age"] = int64(31)

	updated, err := storage.Update(ctx, foundObj)
	require.NoError(t, err)
	assert.True(t, updated)

	// Verify update
	updatedObj, err := storage.FindByID(ctx, "users", "test-user-1")
	require.NoError(t, err)
	require.NotNil(t, updatedObj)
	assert.Equal(t, "updated@example.com", updatedObj.Fields["email"])

	// Test FindByKey
	foundByEmail, err := storage.FindByKey(ctx, "users", "email", "updated@example.com")
	require.NoError(t, err)
	require.NotNil(t, foundByEmail)
	assert.Equal(t, "test-user-1", foundByEmail.ID)

	// Test DeleteByID
	deleted, err := storage.DeleteByID(ctx, "users", "test-user-1")
	require.NoError(t, err)
	assert.True(t, deleted)

	// Verify deletion
	deletedObj, err := storage.FindByID(ctx, "users", "test-user-1")
	require.NoError(t, err)
	assert.Nil(t, deletedObj)
}

func TestScyllaDBFindByKey_InvalidInputs(t *testing.T) {
	storage := &ScyllaDBStorage{}
	ctx := context.Background()

	// Test with empty table name
	_, err := storage.FindByKey(ctx, "", "key", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name, key, and value must not be empty")

	// Test with empty key
	_, err = storage.FindByKey(ctx, "table", "", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name, key, and value must not be empty")

	// Test with empty value
	_, err = storage.FindByKey(ctx, "table", "key", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name, key, and value must not be empty")
}