package main

import (
	"github.com/jadedragon942/ddao/schema"
)

func createExampleSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("restd")

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
	userTable.AddField(schema.ColumnData{
		Name:     "updated_at",
		DataType: "datetime",
		Nullable: true,
		Comment:  "Last update timestamp",
	})

	// Create posts table
	postTable := schema.NewTableSchema("posts")
	postTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Post ID",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "user_id",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "User ID who created the post",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "title",
		DataType: "text",
		Nullable: false,
		Comment:  "Post title",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "content",
		DataType: "text",
		Nullable: false,
		Comment:  "Post content",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "metadata",
		DataType: "json",
		Nullable: true,
		Comment:  "Post metadata",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "published",
		DataType: "boolean",
		Nullable: false,
		Default:  "false",
		Comment:  "Whether the post is published",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})
	postTable.AddField(schema.ColumnData{
		Name:     "updated_at",
		DataType: "datetime",
		Nullable: true,
		Comment:  "Last update timestamp",
	})

	sch.AddTable(userTable)
	sch.AddTable(postTable)
	return sch
}