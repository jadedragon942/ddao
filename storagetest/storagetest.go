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
