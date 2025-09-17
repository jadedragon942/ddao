package infoschema_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jadedragon942/ddao/schema/parser/infoschema"
	_ "github.com/mattn/go-sqlite3"
)

func ExampleParser_ParseSchema() {
	// Open a database connection (using SQLite for this example)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create some sample tables
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE INDEX idx_posts_user_id ON posts(user_id);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}

	// For SQLite, use the SQLiteAdapter since SQLite doesn't have INFORMATION_SCHEMA
	adapter := infoschema.NewSQLiteAdapter(db)
	parsedSchema, err := adapter.ParseSchema("main")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database: %s\n", parsedSchema.DatabaseName)
	fmt.Printf("Tables: %d\n", len(parsedSchema.Tables))

	// Print table information
	for tableName, table := range parsedSchema.Tables {
		fmt.Printf("\nTable: %s\n", tableName)
		fmt.Printf("  Primary Key: %s\n", table.PrimaryKey)
		fmt.Printf("  Fields:\n")
		for fieldName, field := range table.Fields {
			fmt.Printf("    %s: %s (nullable: %v, unique: %v)\n",
				fieldName, field.DataType, field.Nullable, field.Unique)
		}
		if len(table.Indexes) > 0 {
			fmt.Printf("  Indexes: %v\n", table.Indexes)
		}
		if len(table.UniqueKeys) > 0 {
			fmt.Printf("  Unique Keys: %v\n", table.UniqueKeys)
		}
	}
}

func ExampleNewParser() {
	// For databases that support INFORMATION_SCHEMA (MySQL, PostgreSQL, etc.)
	// db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/database")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer db.Close()
	//
	// parser := infoschema.NewParser(db)
	// schema, err := parser.ParseSchema("your_database_name")
	// if err != nil {
	//     log.Fatal(err)
	// }
	//
	// fmt.Printf("Found %d tables\n", len(schema.Tables))

	fmt.Println("Example of using NewParser for INFORMATION_SCHEMA databases")
	// Output: Example of using NewParser for INFORMATION_SCHEMA databases
}