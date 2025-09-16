package tidb

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestTiDBStorage(t *testing.T) {
	connStr := os.Getenv("TIDB_TEST_URL")
	if connStr == "" {
		t.Skip("TIDB_TEST_URL not set, skipping TiDB tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to TiDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestTiDBCRUD(t *testing.T) {
	connStr := os.Getenv("TIDB_TEST_URL")
	if connStr == "" {
		t.Skip("TIDB_TEST_URL not set, skipping TiDB CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to TiDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestTiDBLocal runs tests against a local TiDB instance
// Use: docker run --name tidb-server -d -p 4000:4000 pingcap/tidb:latest
func TestTiDBLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local TiDB test in short mode")
	}

	connStr := "root@tcp(localhost:4000)/test"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local TiDB (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}