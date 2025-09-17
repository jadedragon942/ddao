package main

import (
	"context"
	"log"
	"time"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/orm"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage/scylla"
)

func createUserSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("scylla_example")

	// Create users table
	userTable := schema.NewTableSchema("users")

	// Define fields
	userTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "User ID",
	})

	userTable.AddField(schema.ColumnData{
		Name:     "email",
		DataType: "text",
		Nullable: false,
		Comment:  "User email address",
	})

	userTable.AddField(schema.ColumnData{
		Name:     "age",
		DataType: "integer",
		Nullable: true,
		Comment:  "User age",
	})

	userTable.AddField(schema.ColumnData{
		Name:     "active",
		DataType: "boolean",
		Nullable: false,
		Comment:  "User active status",
	})

	userTable.AddField(schema.ColumnData{
		Name:     "profile",
		DataType: "json",
		Nullable: true,
		Comment:  "User profile data",
	})

	userTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	sch.AddTable(userTable)
	return sch
}

func main() {
	ctx := context.Background()

	// Create schema
	schema := createUserSchema()

	// Initialize ScyllaDB storage
	storage := scylla.New()

	// Connect to ScyllaDB
	// Format: "host:port,host:port/keyspace?consistency=level&timeout=duration"
	// For local development: "localhost:9042/myapp?consistency=one&timeout=5s"
	// For production cluster: "node1:9042,node2:9042,node3:9042/myapp?consistency=quorum&timeout=10s"
	connStr := "localhost:9042/scylla_example?consistency=one&timeout=5s"

	err := storage.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Failed to connect to ScyllaDB:", err)
	}
	defer storage.ResetConnection(ctx)

	// Create ORM instance
	orm := orm.New(schema).WithStorage(storage)

	// Create tables
	err = storage.CreateTables(ctx, schema)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	log.Println("ScyllaDB tables created successfully!")

	// Perform CRUD operations
	performCRUDOperations(ctx, orm)
}

func performCRUDOperations(ctx context.Context, orm *orm.ORM) {
	log.Println("\n=== Starting CRUD Operations ===")

	// CREATE - Insert a new user
	user := object.New()
	user.TableName = "users"
	user.ID = "scylla_user_123"
	user.Fields = map[string]interface{}{
		"email":      "scylla@example.com",
		"age":        int64(30),
		"active":     true,
		"profile":    `{"name": "ScyllaDB User", "location": "Distributed"}`,
		"created_at": time.Now(),
	}

	data, created, err := orm.Insert(ctx, user)
	if err != nil {
		log.Printf("Insert failed: %v", err)
		return
	}
	log.Printf("✅ User created: %t, Data length: %d bytes", created, len(data))

	// READ - Find user by ID
	foundUser, err := orm.FindByID(ctx, "users", "scylla_user_123")
	if err != nil {
		log.Printf("Find by ID failed: %v", err)
		return
	}
	if foundUser != nil {
		email, _ := foundUser.GetString("email")
		age, _ := foundUser.GetInt64("age")
		active, _ := foundUser.GetBool("active")
		log.Printf("✅ Found user: %s, Email: %s, Age: %d, Active: %t", foundUser.ID, email, age, active)
	}

	// READ - Find user by email (note: this requires email to be indexed in production)
	userByEmail, err := orm.FindByKey(ctx, "users", "email", "scylla@example.com")
	if err != nil {
		log.Printf("Find by email failed: %v", err)
		return
	}
	if userByEmail != nil {
		log.Printf("✅ Found user by email: %s", userByEmail.ID)
	}

	// UPDATE - Modify user profile
	user.Fields["profile"] = `{"name": "ScyllaDB User", "location": "Distributed", "updated": true}`
	user.Fields["age"] = int64(31)

	updated, err := orm.Storage.Update(ctx, user)
	if err != nil {
		log.Printf("Update failed: %v", err)
		return
	}
	log.Printf("✅ User updated: %t", updated)

	// Verify update
	updatedUser, err := orm.FindByID(ctx, "users", "scylla_user_123")
	if err != nil {
		log.Printf("Find updated user failed: %v", err)
		return
	}
	if updatedUser != nil {
		updatedAge, _ := updatedUser.GetInt64("age")
		log.Printf("✅ Verified update - new age: %d", updatedAge)
	}

	// CREATE - Insert another user for demonstration
	user2 := object.New()
	user2.TableName = "users"
	user2.ID = "scylla_user_456"
	user2.Fields = map[string]interface{}{
		"email":      "another@example.com",
		"age":        int64(25),
		"active":     false,
		"profile":    `{"name": "Another User", "location": "Cloud"}`,
		"created_at": time.Now(),
	}

	_, created, err = orm.Insert(ctx, user2)
	if err != nil {
		log.Printf("Insert second user failed: %v", err)
		return
	}
	log.Printf("✅ Second user created: %t", created)

	// DELETE - Remove first user
	deleted, err := orm.DeleteByID(ctx, "users", "scylla_user_123")
	if err != nil {
		log.Printf("Delete failed: %v", err)
		return
	}
	log.Printf("✅ User deleted: %t", deleted)

	// Verify deletion
	deletedUser, err := orm.FindByID(ctx, "users", "scylla_user_123")
	if err != nil {
		log.Printf("Find deleted user failed: %v", err)
		return
	}
	if deletedUser == nil {
		log.Println("✅ Confirmed: User was successfully deleted")
	} else {
		log.Println("⚠️ Warning: User still exists after deletion")
	}

	// Clean up - delete second user
	deleted, err = orm.DeleteByID(ctx, "users", "scylla_user_456")
	if err != nil {
		log.Printf("Cleanup delete failed: %v", err)
		return
	}
	log.Printf("✅ Cleanup completed: %t", deleted)

	log.Println("\n=== CRUD Operations Completed Successfully! ===")

	// Note about transactions
	log.Println("\nNote: ScyllaDB does not support traditional ACID transactions.")
	log.Println("For consistency, use lightweight transactions (LWT) with IF conditions")
	log.Println("or implement application-level transaction patterns.")
}