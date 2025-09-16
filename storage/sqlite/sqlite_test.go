package sqlite

import (
	"context"
	"testing"

	"github.com/jadedragon942/ddao/storagetest"
)

func TestSQLiteStorage(t *testing.T) {
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to SQLite storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.StorageTest(t, storage)
}

func TestSQLiteCRUD(t *testing.T) {
	storage := New()
	ctx := context.Background()
	err := storage.Connect(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to SQLite storage: %v", err)
	}
	defer storage.ResetConnection(ctx)

	storagetest.CRUDTest(t, storage)
}
