package main

import (
	"github.com/jadedragon942/ddao/schema"
)

func createWikiSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("wiki")

	// Users table
	userTable := schema.NewTableSchema("users")
	userTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "User ID",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "username",
		DataType: "text",
		Nullable: false,
		Unique:   true,
		Comment:  "Username",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "email",
		DataType: "text",
		Nullable: false,
		Unique:   true,
		Comment:  "Email address",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "password",
		DataType: "text",
		Nullable: false,
		Comment:  "Hashed password",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	// Wiki pages table
	wikiPageTable := schema.NewTableSchema("wiki_pages")
	wikiPageTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Page ID",
	})
	wikiPageTable.AddField(schema.ColumnData{
		Name:     "title",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "Page title",
	})
	wikiPageTable.AddField(schema.ColumnData{
		Name:     "content",
		DataType: "text",
		Nullable: false,
		Comment:  "Page content in markdown",
	})
	wikiPageTable.AddField(schema.ColumnData{
		Name:     "author_id",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "Author user ID",
	})
	wikiPageTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})
	wikiPageTable.AddField(schema.ColumnData{
		Name:     "updated_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Last update timestamp",
	})

	// Sessions table
	sessionTable := schema.NewTableSchema("sessions")
	sessionTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Session ID",
	})
	sessionTable.AddField(schema.ColumnData{
		Name:     "user_id",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "User ID",
	})
	sessionTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})
	sessionTable.AddField(schema.ColumnData{
		Name:     "expires_at",
		DataType: "datetime",
		Nullable: false,
		Index:    true,
		Comment:  "Expiration timestamp",
	})

	sch.AddTable(userTable)
	sch.AddTable(wikiPageTable)
	sch.AddTable(sessionTable)

	return sch
}