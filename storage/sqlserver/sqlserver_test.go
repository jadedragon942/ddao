package sqlserver

import (
	"context"
	"os"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestSQLServerStorage(t *testing.T) {
	connStr := os.Getenv("SQLSERVER_TEST_URL")
	if connStr == "" {
		t.Skip("SQLSERVER_TEST_URL not set, skipping SQL Server tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestSQLServerCRUD(t *testing.T) {
	connStr := os.Getenv("SQLSERVER_TEST_URL")
	if connStr == "" {
		t.Skip("SQLSERVER_TEST_URL not set, skipping SQL Server CRUD tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}

// TestSQLServerLocal runs tests against a local SQL Server instance
// Use: docker run -e "ACCEPT_EULA=Y" -e "SA_PASSWORD=YourStrong@Passw0rd" -p 1433:1433 --name sqlserver-test -d mcr.microsoft.com/mssql/server:2019-latest
func TestSQLServerLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local SQL Server test in short mode")
	}

	connStr := "sqlserver://sa:YourStrong@Passw0rd@localhost:1433?database=tempdb"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local SQL Server (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
	storagetest.CRUDTest(t, storage)
}

func TestSQLServerTransactions(t *testing.T) {
	connStr := os.Getenv("SQLSERVER_TEST_URL")
	if connStr == "" {
		t.Skip("SQLSERVER_TEST_URL not set, skipping SQL Server transaction tests")
	}

	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.TransactionTest(t, storage)
}

func TestSQLServerLocalTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local SQL Server transaction test in short mode")
	}

	connStr := "sqlserver://sa:YourStrong@Passw0rd@localhost:1433?database=tempdb"
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("Failed to connect to local SQL Server (is it running?): %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.TransactionTest(t, storage)
}