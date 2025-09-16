package yugabyte

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestYugabyteDBStorage(t *testing.T) {
	connStr := os.Getenv("YUGABYTE_TEST_URL")
	if connStr == "" {
		t.Skip("YUGABYTE_TEST_URL not set, skipping YugabyteDB tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to YugabyteDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestYugabyteDBCRUD(t *testing.T) {
	connStr := os.Getenv("YUGABYTE_TEST_URL")
	if connStr == "" {
		t.Skip("YUGABYTE_TEST_URL not set, skipping YugabyteDB CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to YugabyteDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestYugabyteDBLocal runs tests against a local YugabyteDB instance
// Use: docker run -d --name yugabyte -p 7000:7000 -p 9000:9000 -p 5433:5433 -p 9042:9042 yugabytedb/yugabyte:latest bin/yugabyted start --daemon=false
func TestYugabyteDBLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local YugabyteDB test in short mode")
	}

	connStr := "postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local YugabyteDB (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}