// Package infoschema provides functionality to parse database schemas using the INFORMATION_SCHEMA standard.
//
// The INFORMATION_SCHEMA is a standardized set of read-only views that provide access to database metadata.
// This package implements a parser that can extract schema information from databases that support the
// INFORMATION_SCHEMA standard, including MySQL, PostgreSQL, SQL Server, and others.
//
// Key features:
//   - Parse complete database schemas or individual tables
//   - Extract table metadata including columns, indexes, constraints
//   - Support for primary keys, unique constraints, and foreign keys
//   - Compatible with the ddao schema package types
//
// Usage:
//
//	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/database")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	parser := infoschema.NewParser(db)
//	schema, err := parser.ParseSchema("mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access parsed schema information
//	for tableName, table := range schema.Tables {
//		fmt.Printf("Table: %s\n", tableName)
//		for fieldName, field := range table.Fields {
//			fmt.Printf("  Field: %s (%s)\n", fieldName, field.DataType)
//		}
//	}
//
// The parser queries the following INFORMATION_SCHEMA views:
//   - TABLES: For table metadata
//   - COLUMNS: For column definitions and properties
//   - STATISTICS: For index information
//   - KEY_COLUMN_USAGE: For key constraints
//   - TABLE_CONSTRAINTS: For constraint types
//   - REFERENTIAL_CONSTRAINTS: For foreign key relationships
package infoschema
