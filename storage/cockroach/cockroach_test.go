package cockroach

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestCockroachDBStorage(t *testing.T) {
	connStr := os.Getenv("COCKROACH_TEST_URL")
	if connStr == "" {
		t.Skip("COCKROACH_TEST_URL not set, skipping CockroachDB tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to CockroachDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestCockroachDBCRUD(t *testing.T) {
	connStr := os.Getenv("COCKROACH_TEST_URL")
	if connStr == "" {
		t.Skip("COCKROACH_TEST_URL not set, skipping CockroachDB CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to CockroachDB storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestCockroachDBLocal runs tests against a local CockroachDB instance
// Use: docker run -d --name=roach --hostname=roach -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start-single-node --insecure
func TestCockroachDBLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local CockroachDB test in short mode")
	}

	connStr := "postgres://root@localhost:26257/defaultdb?sslmode=disable"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local CockroachDB (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}