package storagetest

import (
	"context"
	"testing"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
)

// StorageTest is a simple sanity check for a storage.Storage backend
// This code liberally borrowed and modified from github.com/dgryski/go-shardedkv
func StorageTest(t *testing.T, storage storage.Storage) {
	ctx := context.Background()

	err := storage.CreateTables(ctx, schema.GetTestSchema())
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	v, err := storage.FindByID(ctx, "people", "hello")
	if err != nil {
		t.Errorf("getting a non-existent key was 'ok': v=%v err=%v\n", v, err)
	}

	o := object.New()
	o.TableName = "people"
	o.ID = "hello"
	o.Fields = map[string]any{
		"name":     "John Doe",
		"metadata": "{\"test\":\"wowza\"}",
	}

	_, _, err = storage.Insert(ctx, o)
	if err != nil {
		t.Fatalf("failed inserting a valid key: err=%v\n", err)
	}

	v, err = storage.FindByID(ctx, "people", "hello")
	if err != nil {
		t.Errorf("failed getting a valid key: v=%v err=%v\n", v, err)
	}
	if v == nil {
		t.Fatalf("got nil object for key 'hello'")
	}

	name, _ := v.GetString("name")

	t.Logf("got object: v.ID=%v v.Fields=%+v\n", v.ID, v.Fields)

	if v.ID != "hello" {
		t.Fatalf("got wrong object: v.ID=%v v.Fields=%v\n", v.ID, v.Fields)
	}
	if name != "John Doe" {
		t.Fatalf("got wrong name field: name=%v\n", name)
	}

	var ok bool
	ok, err = storage.DeleteByID(ctx, "people", "hello")
	if ok != true || err != nil {
		t.Fatalf("failed deleting key: ok=%v err=%v\n", ok, err)
	}

	v, err = storage.FindByID(ctx, "people", "hello")
	if v != nil || err != nil {
		t.Fatalf("getting a non-existent key post-delete was 'ok': v=%v err=%v\n", v, err)
	}

	err = storage.ResetConnection(ctx)
	if err != nil {
		t.Fatalf("failed resetting connection for key: err=%v\n", err)
	}
}

// CRUDTest performs comprehensive CRUD testing for a storage.Storage backend
func CRUDTest(t *testing.T, storage storage.Storage) {
	ctx := context.Background()

	// Create tables
	err := storage.CreateTables(ctx, schema.GetTestSchema())
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Test CREATE operations
	testObjects := []*object.Object{
		{
			TableName: "people",
			ID:        "user1",
			Fields: map[string]any{
				"name":     "Alice Johnson",
				"metadata": `{"age": 30, "city": "New York"}`,
			},
		},
		{
			TableName: "people",
			ID:        "user2",
			Fields: map[string]any{
				"name":     "Bob Smith",
				"metadata": `{"age": 25, "city": "Boston"}`,
			},
		},
		{
			TableName: "people",
			ID:        "user3",
			Fields: map[string]any{
				"name":     "Charlie Brown",
				"metadata": `{"age": 35, "city": "Chicago"}`,
			},
		},
	}

	for _, obj := range testObjects {
		_, created, err := storage.Insert(ctx, obj)
		if err != nil {
			t.Errorf("failed to insert object %s: %v", obj.ID, err)
		}
		if !created {
			t.Errorf("object %s was not created", obj.ID)
		}
	}

	// Test READ operations
	// Test FindByID
	obj, err := storage.FindByID(ctx, "people", "user1")
	if err != nil {
		t.Errorf("failed to find object by ID: %v", err)
	}
	if obj == nil {
		t.Fatal("object not found")
	}
	if obj.ID != "user1" {
		t.Errorf("expected ID 'user1', got '%s'", obj.ID)
	}

	name, ok := obj.GetString("name")
	if !ok {
		t.Error("failed to get name field")
	}
	if name != "Alice Johnson" {
		t.Errorf("expected name 'Alice Johnson', got '%s'", name)
	}

	// Test FindByKey
	obj2, err := storage.FindByKey(ctx, "people", "name", "Bob Smith")
	if err != nil {
		t.Errorf("failed to find object by key: %v", err)
	}
	if obj2 == nil {
		t.Error("object not found by name key")
	} else if obj2.ID != "user2" {
		t.Errorf("expected ID 'user2', got '%s'", obj2.ID)
	}

	// Test non-existent object
	nonExistent, err := storage.FindByID(ctx, "people", "nonexistent")
	if err != nil {
		t.Errorf("unexpected error when finding non-existent object: %v", err)
	}
	if nonExistent != nil {
		t.Error("expected nil for non-existent object")
	}

	// Test UPDATE operations
	// Update an existing object
	updateObj := &object.Object{
		TableName: "people",
		ID:        "user1",
		Fields: map[string]any{
			"name":     "Alice Johnson Updated",
			"metadata": `{"age": 31, "city": "San Francisco"}`,
		},
	}

	updated, err := storage.Update(ctx, updateObj)
	if err != nil {
		t.Errorf("failed to update object: %v", err)
	}
	if !updated {
		t.Error("object was not updated")
	}

	// Verify the update
	obj, err = storage.FindByID(ctx, "people", "user1")
	if err != nil {
		t.Errorf("failed to find updated object: %v", err)
	}
	if obj == nil {
		t.Fatal("updated object not found")
	}

	name, ok = obj.GetString("name")
	if !ok {
		t.Error("failed to get updated name field")
	}
	if name != "Alice Johnson Updated" {
		t.Errorf("expected updated name 'Alice Johnson Updated', got '%s'", name)
	}

	// Test updating non-existent object
	nonExistentObj := &object.Object{
		TableName: "people",
		ID:        "nonexistent",
		Fields: map[string]any{
			"name":     "Ghost User",
			"metadata": `{}`,
		},
	}

	updated, err = storage.Update(ctx, nonExistentObj)
	if err != nil {
		t.Errorf("unexpected error when updating non-existent object: %v", err)
	}
	if updated {
		t.Error("expected false when updating non-existent object")
	}

	// Test DELETE operations
	// Delete an existing object
	deleted, err := storage.DeleteByID(ctx, "people", "user2")
	if err != nil {
		t.Errorf("failed to delete object: %v", err)
	}
	if !deleted {
		t.Error("object was not deleted")
	}

	// Verify deletion
	obj, err = storage.FindByID(ctx, "people", "user2")
	if err != nil {
		t.Errorf("unexpected error when finding deleted object: %v", err)
	}
	if obj != nil {
		t.Error("deleted object still exists")
	}

	// Test deleting non-existent object
	deleted, err = storage.DeleteByID(ctx, "people", "nonexistent")
	if err != nil {
		t.Errorf("unexpected error when deleting non-existent object: %v", err)
	}
	if deleted {
		t.Error("expected false when deleting non-existent object")
	}

	// Test UPSERT behavior (Insert or Replace)
	// Insert a new object
	newObj := &object.Object{
		TableName: "people",
		ID:        "user4",
		Fields: map[string]any{
			"name":     "David Wilson",
			"metadata": `{"age": 28, "city": "Seattle"}`,
		},
	}

	_, created, err := storage.Insert(ctx, newObj)
	if err != nil {
		t.Errorf("failed to insert new object: %v", err)
	}
	if !created {
		t.Error("new object was not created")
	}

	// Update the same object via Insert (upsert)
	upsertObj := &object.Object{
		TableName: "people",
		ID:        "user4",
		Fields: map[string]any{
			"name":     "David Wilson Jr",
			"metadata": `{"age": 29, "city": "Portland"}`,
		},
	}

	_, created, err = storage.Insert(ctx, upsertObj)
	if err != nil {
		t.Errorf("failed to upsert object: %v", err)
	}

	// Verify the upsert
	obj, err = storage.FindByID(ctx, "people", "user4")
	if err != nil {
		t.Errorf("failed to find upserted object: %v", err)
	}
	if obj == nil {
		t.Fatal("upserted object not found")
	}

	name, ok = obj.GetString("name")
	if !ok {
		t.Error("failed to get upserted name field")
	}
	if name != "David Wilson Jr" {
		t.Errorf("expected upserted name 'David Wilson Jr', got '%s'", name)
	}

	// Clean up
	err = storage.ResetConnection(ctx)
	if err != nil {
		t.Errorf("failed to reset connection: %v", err)
	}
}

// TransactionTest performs comprehensive transaction testing for a storage.Storage backend
func TransactionTest(t *testing.T, storage storage.Storage) {
	ctx := context.Background()

	// Create tables
	err := storage.CreateTables(ctx, schema.GetTestSchema())
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Test 1: Basic transaction operations - commit
	t.Run("BasicTransactionCommit", func(t *testing.T) {
		tx, err := storage.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "tx_user1",
			Fields: map[string]any{
				"name":     "Transaction User 1",
				"metadata": `{"test": "commit"}`,
			},
		}

		_, created, err := storage.InsertTx(ctx, tx, obj)
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to insert in transaction: %v", err)
		}
		if !created {
			storage.RollbackTx(tx)
			t.Fatal("object was not created in transaction")
		}

		// Object may or may not be visible outside transaction depending on database implementation
		// SQLite in-memory may behave differently than other databases
		foundObj, err := storage.FindByID(ctx, "people", "tx_user1")
		if err != nil {
			// This is expected for some database implementations
			t.Logf("Expected behavior: object not visible outside transaction: %v", err)
		} else if foundObj != nil {
			t.Logf("Note: object visible outside uncommitted transaction (database-specific behavior)")
		}

		// Object should be visible within transaction
		foundObjTx, err := storage.FindByIDTx(ctx, tx, "people", "tx_user1")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to find object in transaction: %v", err)
		}
		if foundObjTx == nil {
			storage.RollbackTx(tx)
			t.Fatal("object not found in transaction")
		}

		// Commit transaction
		err = storage.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit transaction: %v", err)
		}

		// Object should now be visible outside transaction
		foundObj, err = storage.FindByID(ctx, "people", "tx_user1")
		if err != nil {
			t.Fatalf("failed to find committed object: %v", err)
		}
		if foundObj == nil {
			t.Fatal("committed object not found")
		}
		if foundObj.ID != "tx_user1" {
			t.Errorf("expected ID 'tx_user1', got '%s'", foundObj.ID)
		}
	})

	// Test 2: Transaction rollback
	t.Run("TransactionRollback", func(t *testing.T) {
		tx, err := storage.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "tx_user2",
			Fields: map[string]any{
				"name":     "Transaction User 2",
				"metadata": `{"test": "rollback"}`,
			},
		}

		_, created, err := storage.InsertTx(ctx, tx, obj)
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to insert in transaction: %v", err)
		}
		if !created {
			storage.RollbackTx(tx)
			t.Fatal("object was not created in transaction")
		}

		// Object should be visible within transaction
		foundObjTx, err := storage.FindByIDTx(ctx, tx, "people", "tx_user2")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to find object in transaction: %v", err)
		}
		if foundObjTx == nil {
			storage.RollbackTx(tx)
			t.Fatal("object not found in transaction")
		}

		// Rollback transaction
		err = storage.RollbackTx(tx)
		if err != nil {
			t.Fatalf("failed to rollback transaction: %v", err)
		}

		// Object should not be visible after rollback
		foundObj, err := storage.FindByID(ctx, "people", "tx_user2")
		if err != nil {
			t.Fatalf("unexpected error finding rolled back object: %v", err)
		}
		if foundObj != nil {
			t.Error("rolled back object should not be visible")
		}
	})

	// Test 3: Transaction update operations
	t.Run("TransactionUpdate", func(t *testing.T) {
		// First insert an object normally
		obj := &object.Object{
			TableName: "people",
			ID:        "tx_user3",
			Fields: map[string]any{
				"name":     "Transaction User 3",
				"metadata": `{"test": "original"}`,
			},
		}

		_, created, err := storage.Insert(ctx, obj)
		if err != nil {
			t.Fatalf("failed to insert initial object: %v", err)
		}
		if !created {
			t.Fatal("initial object was not created")
		}

		// Start transaction and update
		tx, err := storage.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		updateObj := &object.Object{
			TableName: "people",
			ID:        "tx_user3",
			Fields: map[string]any{
				"name":     "Transaction User 3 Updated",
				"metadata": `{"test": "updated"}`,
			},
		}

		updated, err := storage.UpdateTx(ctx, tx, updateObj)
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to update in transaction: %v", err)
		}
		if !updated {
			storage.RollbackTx(tx)
			t.Fatal("object was not updated in transaction")
		}

		// Verify update within transaction
		foundObjTx, err := storage.FindByIDTx(ctx, tx, "people", "tx_user3")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to find updated object in transaction: %v", err)
		}
		if foundObjTx == nil {
			storage.RollbackTx(tx)
			t.Fatal("updated object not found in transaction")
		}

		name, ok := foundObjTx.GetString("name")
		if !ok {
			storage.RollbackTx(tx)
			t.Fatal("failed to get name field")
		}
		if name != "Transaction User 3 Updated" {
			storage.RollbackTx(tx)
			t.Errorf("expected updated name 'Transaction User 3 Updated', got '%s'", name)
		}

		// Original object may still be visible outside transaction depending on database implementation
		foundObj, err := storage.FindByID(ctx, "people", "tx_user3")
		if err != nil {
			t.Logf("Expected behavior: original object not accessible outside transaction: %v", err)
		} else if foundObj != nil {
			origName, ok := foundObj.GetString("name")
			if ok && origName == "Transaction User 3" {
				t.Logf("Note: original object still visible outside transaction (database-specific behavior)")
			}
		}

		// Commit transaction
		err = storage.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit transaction: %v", err)
		}

		// Verify update is now visible outside transaction
		foundObj, err = storage.FindByID(ctx, "people", "tx_user3")
		if err != nil {
			t.Fatalf("failed to find committed updated object: %v", err)
		}
		if foundObj == nil {
			t.Fatal("committed updated object not found")
		}

		name, ok = foundObj.GetString("name")
		if !ok {
			t.Fatal("failed to get committed name field")
		}
		if name != "Transaction User 3 Updated" {
			t.Errorf("expected committed name 'Transaction User 3 Updated', got '%s'", name)
		}
	})

	// Test 4: Transaction delete operations
	t.Run("TransactionDelete", func(t *testing.T) {
		// First insert an object normally
		obj := &object.Object{
			TableName: "people",
			ID:        "tx_user4",
			Fields: map[string]any{
				"name":     "Transaction User 4",
				"metadata": `{"test": "delete"}`,
			},
		}

		_, created, err := storage.Insert(ctx, obj)
		if err != nil {
			t.Fatalf("failed to insert initial object: %v", err)
		}
		if !created {
			t.Fatal("initial object was not created")
		}

		// Start transaction and delete
		tx, err := storage.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		deleted, err := storage.DeleteByIDTx(ctx, tx, "people", "tx_user4")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to delete in transaction: %v", err)
		}
		if !deleted {
			storage.RollbackTx(tx)
			t.Fatal("object was not deleted in transaction")
		}

		// Object should not be visible within transaction
		foundObjTx, err := storage.FindByIDTx(ctx, tx, "people", "tx_user4")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("unexpected error finding deleted object in transaction: %v", err)
		}
		if foundObjTx != nil {
			storage.RollbackTx(tx)
			t.Error("deleted object should not be visible in transaction")
		}

		// Object may still be visible outside transaction depending on database implementation
		foundObj, err := storage.FindByID(ctx, "people", "tx_user4")
		if err != nil {
			t.Logf("Expected behavior: object not accessible outside transaction: %v", err)
		} else if foundObj == nil {
			t.Logf("Note: object already deleted outside transaction (database-specific behavior)")
		} else {
			t.Logf("Note: original object still visible outside transaction (database-specific behavior)")
		}

		// Commit transaction
		err = storage.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit transaction: %v", err)
		}

		// Object should now be deleted outside transaction
		foundObj, err = storage.FindByID(ctx, "people", "tx_user4")
		if err != nil {
			t.Fatalf("unexpected error finding deleted object: %v", err)
		}
		if foundObj != nil {
			t.Error("committed deleted object should not be visible")
		}
	})

	// Test 5: Transaction with FindByKey operations
	t.Run("TransactionFindByKey", func(t *testing.T) {
		tx, err := storage.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "tx_user5",
			Fields: map[string]any{
				"name":     "Transaction User 5",
				"metadata": `{"test": "findbykey"}`,
			},
		}

		_, created, err := storage.InsertTx(ctx, tx, obj)
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to insert in transaction: %v", err)
		}
		if !created {
			storage.RollbackTx(tx)
			t.Fatal("object was not created in transaction")
		}

		// Find by key within transaction
		foundObjTx, err := storage.FindByKeyTx(ctx, tx, "people", "name", "Transaction User 5")
		if err != nil {
			storage.RollbackTx(tx)
			t.Fatalf("failed to find object by key in transaction: %v", err)
		}
		if foundObjTx == nil {
			storage.RollbackTx(tx)
			t.Fatal("object not found by key in transaction")
		}
		if foundObjTx.ID != "tx_user5" {
			storage.RollbackTx(tx)
			t.Errorf("expected ID 'tx_user5', got '%s'", foundObjTx.ID)
		}

		// Commit transaction
		err = storage.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit transaction: %v", err)
		}

		// Find by key outside transaction
		foundObj, err := storage.FindByKey(ctx, "people", "name", "Transaction User 5")
		if err != nil {
			t.Fatalf("failed to find committed object by key: %v", err)
		}
		if foundObj == nil {
			t.Fatal("committed object not found by key")
		}
		if foundObj.ID != "tx_user5" {
			t.Errorf("expected committed ID 'tx_user5', got '%s'", foundObj.ID)
		}
	})

	// Test 6: Error handling - nil transaction
	t.Run("NilTransactionHandling", func(t *testing.T) {
		obj := &object.Object{
			TableName: "people",
			ID:        "nil_tx_test",
			Fields: map[string]any{
				"name":     "Nil TX Test",
				"metadata": `{"test": "nil"}`,
			},
		}

		// Test InsertTx with nil transaction
		_, _, err := storage.InsertTx(ctx, nil, obj)
		if err == nil {
			t.Error("expected error when using nil transaction for InsertTx")
		}

		// Test UpdateTx with nil transaction
		_, err = storage.UpdateTx(ctx, nil, obj)
		if err == nil {
			t.Error("expected error when using nil transaction for UpdateTx")
		}

		// Test FindByIDTx with nil transaction
		_, err = storage.FindByIDTx(ctx, nil, "people", "nil_tx_test")
		if err == nil {
			t.Error("expected error when using nil transaction for FindByIDTx")
		}

		// Test FindByKeyTx with nil transaction
		_, err = storage.FindByKeyTx(ctx, nil, "people", "name", "Nil TX Test")
		if err == nil {
			t.Error("expected error when using nil transaction for FindByKeyTx")
		}

		// Test DeleteByIDTx with nil transaction
		_, err = storage.DeleteByIDTx(ctx, nil, "people", "nil_tx_test")
		if err == nil {
			t.Error("expected error when using nil transaction for DeleteByIDTx")
		}

		// Test CommitTx with nil transaction
		err = storage.CommitTx(nil)
		if err == nil {
			t.Error("expected error when committing nil transaction")
		}

		// Test RollbackTx with nil transaction
		err = storage.RollbackTx(nil)
		if err == nil {
			t.Error("expected error when rolling back nil transaction")
		}
	})

	// Clean up
	err = storage.ResetConnection(ctx)
	if err != nil {
		t.Errorf("failed to reset connection: %v", err)
	}
}
