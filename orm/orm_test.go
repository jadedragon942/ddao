package orm

import (
	"context"
	"testing"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	sqliteStorage "github.com/jadedragon942/ddao/storage/sqlite"
)

func getTestSchema() *schema.Schema {
	sch := schema.New()

	table := schema.NewTableSchema("people")
	idField := schema.ColumnData{
		Name:          "id",
		DataType:      "text",
		Nullable:      false,
		Default:       nil,
		Comment:       "Primary key for the person",
		Unique:        true,
		Index:         true,
		AutoIncrement: false,
	}
	table.AddField(idField)
	nameField := schema.ColumnData{
		Name:          "name",
		DataType:      "text",
		Nullable:      false,
		Default:       nil,
		Comment:       "Name of the person",
		Unique:        false,
		Index:         true,
		AutoIncrement: false,
	}
	table.AddField(nameField)

	metadataField := schema.ColumnData{
		Name:     "metadata",
		DataType: "json",
		Nullable: true,
		Default:  nil,
		Comment:  "Additional metadata for the person",
	}

	table.AddField(metadataField)

	sch.AddTable(table)

	return sch
}

func TestORM(t *testing.T) {
	sch := getTestSchema()
	o := New(sch)
	o.WithStorage(sqliteStorage.New())

}

func TestORMTransactions(t *testing.T) {
	sch := getTestSchema()
	o := New(sch)
	storage := sqliteStorage.New()
	o.WithStorage(storage)

	ctx := context.Background()
	err := o.Connect(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to storage: %v", err)
	}
	defer o.ResetConnection(ctx)

	err = storage.CreateTables(ctx, sch)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test basic transaction operations through ORM
	t.Run("ORMTransactionCommit", func(t *testing.T) {
		tx, err := o.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user1",
			Fields: map[string]any{
				"name":     "ORM Transaction User 1",
				"metadata": `{"test": "orm_commit"}`,
			},
		}

		_, created, err := o.InsertTx(ctx, tx, obj)
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to insert via ORM transaction: %v", err)
		}
		if !created {
			o.RollbackTx(tx)
			t.Fatal("object was not created via ORM transaction")
		}

		// Verify object exists within transaction
		foundObj, err := o.FindByIDTx(ctx, tx, "people", "orm_tx_user1")
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to find object via ORM transaction: %v", err)
		}
		if foundObj == nil {
			o.RollbackTx(tx)
			t.Fatal("object not found via ORM transaction")
		}

		err = o.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit ORM transaction: %v", err)
		}

		// Verify object exists after commit
		foundObj, err = o.FindByID(ctx, "people", "orm_tx_user1")
		if err != nil {
			t.Fatalf("failed to find committed object via ORM: %v", err)
		}
		if foundObj == nil {
			t.Fatal("committed object not found via ORM")
		}
	})

	t.Run("ORMTransactionRollback", func(t *testing.T) {
		tx, err := o.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user2",
			Fields: map[string]any{
				"name":     "ORM Transaction User 2",
				"metadata": `{"test": "orm_rollback"}`,
			},
		}

		_, created, err := o.InsertTx(ctx, tx, obj)
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to insert via ORM transaction: %v", err)
		}
		if !created {
			o.RollbackTx(tx)
			t.Fatal("object was not created via ORM transaction")
		}

		err = o.RollbackTx(tx)
		if err != nil {
			t.Fatalf("failed to rollback ORM transaction: %v", err)
		}

		// Verify object does not exist after rollback
		foundObj, err := o.FindByID(ctx, "people", "orm_tx_user2")
		if err != nil {
			t.Fatalf("unexpected error finding rolled back object via ORM: %v", err)
		}
		if foundObj != nil {
			t.Error("rolled back object should not be visible via ORM")
		}
	})

	t.Run("ORMTransactionUpdate", func(t *testing.T) {
		// First insert an object normally
		obj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user3",
			Fields: map[string]any{
				"name":     "ORM Transaction User 3",
				"metadata": `{"test": "orm_original"}`,
			},
		}

		_, created, err := o.Insert(ctx, obj)
		if err != nil {
			t.Fatalf("failed to insert initial object via ORM: %v", err)
		}
		if !created {
			t.Fatal("initial object was not created via ORM")
		}

		// Start transaction and update
		tx, err := o.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		updateObj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user3",
			Fields: map[string]any{
				"name":     "ORM Transaction User 3 Updated",
				"metadata": `{"test": "orm_updated"}`,
			},
		}

		updated, err := o.UpdateTx(ctx, tx, updateObj)
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to update via ORM transaction: %v", err)
		}
		if !updated {
			o.RollbackTx(tx)
			t.Fatal("object was not updated via ORM transaction")
		}

		err = o.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit ORM transaction: %v", err)
		}

		// Verify update is visible after commit
		foundObj, err := o.FindByID(ctx, "people", "orm_tx_user3")
		if err != nil {
			t.Fatalf("failed to find updated object via ORM: %v", err)
		}
		if foundObj == nil {
			t.Fatal("updated object not found via ORM")
		}

		name, ok := foundObj.GetString("name")
		if !ok {
			t.Fatal("failed to get name field from updated object")
		}
		if name != "ORM Transaction User 3 Updated" {
			t.Errorf("expected updated name 'ORM Transaction User 3 Updated', got '%s'", name)
		}
	})

	t.Run("ORMTransactionDelete", func(t *testing.T) {
		// First insert an object normally
		obj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user4",
			Fields: map[string]any{
				"name":     "ORM Transaction User 4",
				"metadata": `{"test": "orm_delete"}`,
			},
		}

		_, created, err := o.Insert(ctx, obj)
		if err != nil {
			t.Fatalf("failed to insert initial object via ORM: %v", err)
		}
		if !created {
			t.Fatal("initial object was not created via ORM")
		}

		// Start transaction and delete
		tx, err := o.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		deleted, err := o.DeleteByIDTx(ctx, tx, "people", "orm_tx_user4")
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to delete via ORM transaction: %v", err)
		}
		if !deleted {
			o.RollbackTx(tx)
			t.Fatal("object was not deleted via ORM transaction")
		}

		err = o.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit ORM transaction: %v", err)
		}

		// Verify object is deleted after commit
		foundObj, err := o.FindByID(ctx, "people", "orm_tx_user4")
		if err != nil {
			t.Fatalf("unexpected error finding deleted object via ORM: %v", err)
		}
		if foundObj != nil {
			t.Error("deleted object should not be visible via ORM")
		}
	})

	t.Run("ORMTransactionFindByKey", func(t *testing.T) {
		tx, err := o.BeginTx(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		obj := &object.Object{
			TableName: "people",
			ID:        "orm_tx_user5",
			Fields: map[string]any{
				"name":     "ORM Transaction User 5",
				"metadata": `{"test": "orm_findbykey"}`,
			},
		}

		_, created, err := o.InsertTx(ctx, tx, obj)
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to insert via ORM transaction: %v", err)
		}
		if !created {
			o.RollbackTx(tx)
			t.Fatal("object was not created via ORM transaction")
		}

		// Find by key within transaction
		foundObj, err := o.FindByKeyTx(ctx, tx, "people", "name", "ORM Transaction User 5")
		if err != nil {
			o.RollbackTx(tx)
			t.Fatalf("failed to find object by key via ORM transaction: %v", err)
		}
		if foundObj == nil {
			o.RollbackTx(tx)
			t.Fatal("object not found by key via ORM transaction")
		}
		if foundObj.ID != "orm_tx_user5" {
			o.RollbackTx(tx)
			t.Errorf("expected ID 'orm_tx_user5', got '%s'", foundObj.ID)
		}

		err = o.CommitTx(tx)
		if err != nil {
			t.Fatalf("failed to commit ORM transaction: %v", err)
		}

		// Find by key outside transaction
		foundObj, err = o.FindByKey(ctx, "people", "name", "ORM Transaction User 5")
		if err != nil {
			t.Fatalf("failed to find committed object by key via ORM: %v", err)
		}
		if foundObj == nil {
			t.Fatal("committed object not found by key via ORM")
		}
		if foundObj.ID != "orm_tx_user5" {
			t.Errorf("expected committed ID 'orm_tx_user5', got '%s'", foundObj.ID)
		}
	})
}
