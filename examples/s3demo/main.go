package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/orm"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage/s3"
)

func main() {
	var (
		connStr = flag.String("conn", "s3://ddao-demo-bucket/demo-data?region=us-east-1&endpoint=http://localhost:9000", "S3 connection string")
		verbose = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	ctx := context.Background()

	// Create schema
	schema := createDemoSchema()

	// Initialize S3 storage
	storage := s3.New()
	err := storage.Connect(ctx, *connStr)
	if err != nil {
		log.Fatalf("Failed to connect to S3: %v", err)
	}
	defer storage.ResetConnection(ctx)

	// Create tables
	err = storage.CreateTables(ctx, schema)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create ORM
	ormInstance := orm.New(schema).WithStorage(storage)

	if *verbose {
		log.Println("S3 storage initialized successfully")
	}

	// Perform demo operations
	performDemoOperations(ctx, ormInstance, *verbose)
}

func createDemoSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("s3demo")

	// Create users table
	userTable := schema.NewTableSchema("users")
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
		Unique:   true,
		Comment:  "User email address",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "name",
		DataType: "text",
		Nullable: false,
		Comment:  "User full name",
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

	// Create documents table
	docTable := schema.NewTableSchema("documents")
	docTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Document ID",
	})
	docTable.AddField(schema.ColumnData{
		Name:     "user_id",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "User ID who owns the document",
	})
	docTable.AddField(schema.ColumnData{
		Name:     "title",
		DataType: "text",
		Nullable: false,
		Comment:  "Document title",
	})
	docTable.AddField(schema.ColumnData{
		Name:     "content",
		DataType: "text",
		Nullable: false,
		Comment:  "Document content",
	})
	docTable.AddField(schema.ColumnData{
		Name:     "metadata",
		DataType: "json",
		Nullable: true,
		Comment:  "Document metadata",
	})
	docTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	sch.AddTable(userTable)
	sch.AddTable(docTable)
	return sch
}

func performDemoOperations(ctx context.Context, orm *orm.ORM, verbose bool) {
	log.Println("=== S3 Storage Demo ===")

	// Create a user
	user := object.New()
	user.TableName = "users"
	user.ID = "user123"
	user.Fields = map[string]interface{}{
		"email":      "john.doe@example.com",
		"name":       "John Doe",
		"profile":    `{"age": 30, "city": "New York", "interests": ["technology", "reading"]}`,
		"created_at": time.Now().Format(time.RFC3339),
	}

	log.Println("Creating user...")
	data, created, err := orm.Insert(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	if verbose {
		log.Printf("User created: %t, Data size: %d bytes", created, len(data))
	}

	// Find user by ID
	log.Println("Finding user by ID...")
	foundUser, err := orm.FindByID(ctx, "users", "user123")
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}
	if foundUser != nil {
		log.Printf("Found user: %s (%s)", foundUser.Fields["name"], foundUser.Fields["email"])
	}

	// Find user by email
	log.Println("Finding user by email...")
	userByEmail, err := orm.FindByKey(ctx, "users", "email", "john.doe@example.com")
	if err != nil {
		log.Fatalf("Failed to find user by email: %v", err)
	}
	if userByEmail != nil {
		log.Printf("Found user by email: %s", userByEmail.ID)
	}

	// Create some documents
	documents := []struct {
		id      string
		title   string
		content string
	}{
		{"doc1", "My First Document", "This is the content of my first document stored in S3."},
		{"doc2", "S3 Storage Guide", "A comprehensive guide on using S3 for object storage."},
		{"doc3", "DDAO Tutorial", "Learn how to use DDAO with different storage backends."},
	}

	log.Println("Creating documents...")
	for _, d := range documents {
		doc := object.New()
		doc.TableName = "documents"
		doc.ID = d.id
		doc.Fields = map[string]interface{}{
			"user_id":    "user123",
			"title":      d.title,
			"content":    d.content,
			"metadata":   `{"format": "text", "language": "en"}`,
			"created_at": time.Now().Format(time.RFC3339),
		}

		_, created, err := orm.Insert(ctx, doc)
		if err != nil {
			log.Printf("Failed to create document %s: %v", d.id, err)
			continue
		}
		if verbose {
			log.Printf("Document %s created: %t", d.id, created)
		}
	}

	// Find documents by user
	log.Println("Finding documents by user...")
	doc, err := orm.FindByKey(ctx, "documents", "user_id", "user123")
	if err != nil {
		log.Printf("Failed to find document: %v", err)
	} else if doc != nil {
		log.Printf("Found document: %s", doc.Fields["title"])
	}

	// Update user profile
	log.Println("Updating user profile...")
	user.Fields["profile"] = `{"age": 31, "city": "San Francisco", "interests": ["technology", "reading", "travel"]}`
	updated, err := orm.Storage.Update(ctx, user)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	log.Printf("User updated: %t", updated)

	// Verify update
	updatedUser, err := orm.FindByID(ctx, "users", "user123")
	if err != nil {
		log.Fatalf("Failed to find updated user: %v", err)
	}
	if updatedUser != nil {
		log.Printf("Updated user profile: %s", updatedUser.Fields["profile"])
	}

	// Clean up (optional)
	if verbose {
		log.Println("Cleaning up...")
		for _, d := range documents {
			deleted, err := orm.DeleteByID(ctx, "documents", d.id)
			if err != nil {
				log.Printf("Failed to delete document %s: %v", d.id, err)
			} else {
				log.Printf("Document %s deleted: %t", d.id, deleted)
			}
		}

		deleted, err := orm.DeleteByID(ctx, "users", "user123")
		if err != nil {
			log.Printf("Failed to delete user: %v", err)
		} else {
			log.Printf("User deleted: %t", deleted)
		}
	}

	log.Println("=== Demo completed ===")
}