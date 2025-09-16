package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestPostgreSQLStorage(t *testing.T) {
	connStr := os.Getenv("POSTGRES_TEST_URL")
	if connStr == "" {
		t.Skip("POSTGRES_TEST_URL not set, skipping PostgreSQL tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestPostgreSQLCRUD(t *testing.T) {
	connStr := os.Getenv("POSTGRES_TEST_URL")
	if connStr == "" {
		t.Skip("POSTGRES_TEST_URL not set, skipping PostgreSQL CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestPostgreSQLLocal runs tests against a local PostgreSQL instance
// Use: docker run --name postgres-test -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=testdb -p 5432:5432 -d postgres:13
func TestPostgreSQLLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local PostgreSQL test in short mode")
	}

	connStr := "postgres://postgres:testpass@localhost:5432/testdb?sslmode=disable"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local PostgreSQL (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}