package s3

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBucket = "ddao-test-bucket"
	testPrefix = "test-data"
)

// getTestConnectionString returns a connection string for testing
// It supports both AWS S3 and MinIO (local S3-compatible storage)
func getTestConnectionString() string {
	// Check for MinIO test setup first
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		return fmt.Sprintf("s3://%s/%s?region=us-east-1&endpoint=%s&verbose=true", testBucket, testPrefix, endpoint)
	}

	// Check for AWS S3 test setup
	if bucket := os.Getenv("S3_TEST_BUCKET"); bucket != "" {
		region := os.Getenv("S3_TEST_REGION")
		if region == "" {
			region = "us-east-1"
		}
		return fmt.Sprintf("s3://%s/%s?region=%s&verbose=true", bucket, testPrefix, region)
	}

	// Default to MinIO on localhost (for local testing)
	return fmt.Sprintf("s3://%s/%s?region=us-east-1&endpoint=http://localhost:9000&verbose=true", testBucket, testPrefix)
}

func createTestStorage(t *testing.T) *S3Storage {
	storage := New().(*S3Storage)

	ctx := context.Background()
	connStr := getTestConnectionString()

	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Could not connect to S3: %v (use MINIO_ENDPOINT or S3_TEST_BUCKET env vars for testing)", err)
	}

	return storage
}

func createTestSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("testdb")

	// Create a test table
	userTable := schema.NewTableSchema("users")
	userTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "User ID",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "email",
		DataType: "text",
		Nullable: false,
		Unique:   true,
		Comment:  "User email",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "name",
		DataType: "text",
		Nullable: false,
		Comment:  "User name",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "profile",
		DataType: "json",
		Nullable: true,
		Comment:  "User profile data",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	sch.AddTable(userTable)
	return sch
}

func TestS3Storage_Connect(t *testing.T) {
	tests := []struct {
		name    string
		connStr string
		wantErr bool
	}{
		{
			name:    "valid S3 connection string",
			connStr: "s3://test-bucket/test-prefix?region=us-east-1",
			wantErr: true, // Will fail because bucket likely doesn't exist
		},
		{
			name:    "invalid scheme",
			connStr: "http://test-bucket/test-prefix",
			wantErr: true,
		},
		{
			name:    "missing bucket",
			connStr: "s3:///test-prefix?region=us-east-1",
			wantErr: true,
		},
		{
			name:    "valid MinIO connection string",
			connStr: "s3://test-bucket/test-prefix?region=us-east-1&endpoint=http://localhost:9000",
			wantErr: true, // Will fail because MinIO likely isn't running
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := New().(*S3Storage)
			err := storage.Connect(context.Background(), tt.connStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_CreateTables(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Verify schema was uploaded
	assert.NotNil(t, storage.sch)
	assert.Equal(t, "testdb", storage.sch.DatabaseName)
}

func TestS3Storage_BasicCRUD(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	schema := createTestSchema()
	ctx := context.Background()

	// Create tables
	err := storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Test Insert
	user := object.New()
	user.TableName = "users"
	user.ID = "user123"
	user.Fields = map[string]interface{}{
		"email":      "test@example.com",
		"name":       "Test User",
		"profile":    `{"age": 30, "city": "New York"}`,
		"created_at": time.Now().Format(time.RFC3339),
	}

	data, created, err := storage.Insert(ctx, user)
	require.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, data)

	// Test FindByID
	foundUser, err := storage.FindByID(ctx, "users", "user123")
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, "user123", foundUser.ID)
	assert.Equal(t, "test@example.com", foundUser.Fields["email"])

	// Test FindByKey
	userByEmail, err := storage.FindByKey(ctx, "users", "email", "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, userByEmail)
	assert.Equal(t, "user123", userByEmail.ID)

	// Test Update
	user.Fields["name"] = "Updated Test User"
	updated, err := storage.Update(ctx, user)
	require.NoError(t, err)
	assert.True(t, updated)

	// Verify update
	updatedUser, err := storage.FindByID(ctx, "users", "user123")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test User", updatedUser.Fields["name"])

	// Test Delete
	deleted, err := storage.DeleteByID(ctx, "users", "user123")
	require.NoError(t, err)
	assert.True(t, deleted)

	// Verify deletion
	deletedUser, err := storage.FindByID(ctx, "users", "user123")
	require.NoError(t, err)
	assert.Nil(t, deletedUser)
}

func TestS3Storage_FindByKeyNotFound(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Search for non-existent user
	user, err := storage.FindByKey(ctx, "users", "email", "nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestS3Storage_TransactionMethods(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Test transaction methods (note: S3 doesn't have real transactions)
	tx, err := storage.BeginTx(ctx)
	require.NoError(t, err)
	assert.NotNil(t, tx)

	// Test InsertTx
	user := object.New()
	user.TableName = "users"
	user.ID = "txuser123"
	user.Fields = map[string]interface{}{
		"email":      "tx@example.com",
		"name":       "Transaction User",
		"created_at": time.Now().Format(time.RFC3339),
	}

	data, created, err := storage.InsertTx(ctx, tx, user)
	require.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, data)

	// Test FindByIDTx
	foundUser, err := storage.FindByIDTx(ctx, tx, "users", "txuser123")
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, "txuser123", foundUser.ID)

	// Test UpdateTx
	user.Fields["name"] = "Updated Transaction User"
	updated, err := storage.UpdateTx(ctx, tx, user)
	require.NoError(t, err)
	assert.True(t, updated)

	// Test FindByKeyTx
	userByEmail, err := storage.FindByKeyTx(ctx, tx, "users", "email", "tx@example.com")
	require.NoError(t, err)
	require.NotNil(t, userByEmail)
	assert.Equal(t, "txuser123", userByEmail.ID)

	// Test DeleteByIDTx
	deleted, err := storage.DeleteByIDTx(ctx, tx, "users", "txuser123")
	require.NoError(t, err)
	assert.True(t, deleted)

	// Commit transaction
	err = storage.CommitTx(tx)
	require.NoError(t, err)
}

func TestS3Storage_MultipleObjects(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	require.NoError(t, err)

	// Create multiple users
	users := []struct {
		id    string
		email string
		name  string
	}{
		{"user1", "user1@example.com", "User One"},
		{"user2", "user2@example.com", "User Two"},
		{"user3", "user3@example.com", "User Three"},
	}

	// Insert all users
	for _, u := range users {
		user := object.New()
		user.TableName = "users"
		user.ID = u.id
		user.Fields = map[string]interface{}{
			"email":      u.email,
			"name":       u.name,
			"created_at": time.Now().Format(time.RFC3339),
		}

		_, created, err := storage.Insert(ctx, user)
		require.NoError(t, err)
		assert.True(t, created)
	}

	// Test finding each user
	for _, u := range users {
		// Find by ID
		foundUser, err := storage.FindByID(ctx, "users", u.id)
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, u.id, foundUser.ID)
		assert.Equal(t, u.email, foundUser.Fields["email"])

		// Find by email
		userByEmail, err := storage.FindByKey(ctx, "users", "email", u.email)
		require.NoError(t, err)
		require.NotNil(t, userByEmail)
		assert.Equal(t, u.id, userByEmail.ID)
	}

	// Clean up
	for _, u := range users {
		deleted, err := storage.DeleteByID(ctx, "users", u.id)
		require.NoError(t, err)
		assert.True(t, deleted)
	}
}

// TestS3Storage_StorageTest runs the standard DDAO storage tests
func TestS3Storage_StorageTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping S3 storage test in short mode")
	}

	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	// Run the standard storage tests
	storagetest.StorageTest(t, storage)
}

// TestS3Storage_CRUDTest runs the standard DDAO CRUD tests
func TestS3Storage_CRUDTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping S3 CRUD test in short mode")
	}

	storage := createTestStorage(t)
	defer storage.ResetConnection(context.Background())

	// Run the standard CRUD tests
	storagetest.CRUDTest(t, storage)
}

// BenchmarkS3Storage_Insert benchmarks the insert operation
func BenchmarkS3Storage_Insert(b *testing.B) {
	storage := createTestStorage(&testing.T{})
	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	if err != nil {
		b.Fatalf("Failed to create tables: %v", err)
	}
	defer storage.ResetConnection(ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		user := object.New()
		user.TableName = "users"
		user.ID = fmt.Sprintf("benchuser%d", i)
		user.Fields = map[string]interface{}{
			"email":      fmt.Sprintf("bench%d@example.com", i),
			"name":       fmt.Sprintf("Bench User %d", i),
			"created_at": time.Now().Format(time.RFC3339),
		}

		_, _, err := storage.Insert(ctx, user)
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

// BenchmarkS3Storage_FindByID benchmarks the find by ID operation
func BenchmarkS3Storage_FindByID(b *testing.B) {
	storage := createTestStorage(&testing.T{})
	schema := createTestSchema()
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema)
	if err != nil {
		b.Fatalf("Failed to create tables: %v", err)
	}
	defer storage.ResetConnection(ctx)

	// Create test data
	user := object.New()
	user.TableName = "users"
	user.ID = "benchuser"
	user.Fields = map[string]interface{}{
		"email":      "bench@example.com",
		"name":       "Bench User",
		"created_at": time.Now().Format(time.RFC3339),
	}

	_, _, err = storage.Insert(ctx, user)
	if err != nil {
		b.Fatalf("Failed to insert test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := storage.FindByID(ctx, "users", "benchuser")
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}
	}
}