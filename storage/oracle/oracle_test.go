package oracle

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestOracleStorage(t *testing.T) {
	connStr := os.Getenv("ORACLE_TEST_URL")
	if connStr == "" {
		t.Skip("ORACLE_TEST_URL not set, skipping Oracle tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestOracleCRUD(t *testing.T) {
	connStr := os.Getenv("ORACLE_TEST_URL")
	if connStr == "" {
		t.Skip("ORACLE_TEST_URL not set, skipping Oracle CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestOracleLocal runs tests against a local Oracle instance
// Use: docker run -d -p 1521:1521 -e ORACLE_PASSWORD=OraclePassword123 gvenzl/oracle-xe:21-slim
func TestOracleLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local Oracle test in short mode")
	}

	connStr := "oracle://system:OraclePassword123@localhost:1521/xe"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local Oracle (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}

func TestOracleTransactions(t *testing.T) {
	connStr := os.Getenv("ORACLE_TEST_URL")
	if connStr == "" {
		t.Skip("ORACLE_TEST_URL not set, skipping Oracle transaction tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.TransactionTest(t, storage)
}

func TestOracleLocalTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local Oracle transaction test in short mode")
	}

	connStr := "oracle://system:OraclePassword123@localhost:1521/xe"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local Oracle (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.TransactionTest(t, storage)
}