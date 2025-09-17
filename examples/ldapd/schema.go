package main

import (
	"github.com/jadedragon942/ddao/schema"
)

func createLDAPSchema() *schema.Schema {
	sch := schema.New()
	sch.SetDatabaseName("ldapd")

	// Create entries table for LDAP directory entries
	entryTable := schema.NewTableSchema("entries")
	entryTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Distinguished Name (DN) - unique identifier",
	})
	entryTable.AddField(schema.ColumnData{
		Name:     "parent_dn",
		DataType: "text",
		Nullable: true,
		Index:    true,
		Comment:  "Parent DN for hierarchical structure",
	})
	entryTable.AddField(schema.ColumnData{
		Name:     "object_class",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "LDAP object class (e.g., person, organizationalUnit)",
	})
	entryTable.AddField(schema.ColumnData{
		Name:     "attributes",
		DataType: "json",
		Nullable: false,
		Comment:  "LDAP attributes as JSON",
	})
	entryTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})
	entryTable.AddField(schema.ColumnData{
		Name:     "updated_at",
		DataType: "datetime",
		Nullable: true,
		Comment:  "Last update timestamp",
	})

	// Create users table for authentication
	userTable := schema.NewTableSchema("users")
	userTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "User DN",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "password_hash",
		DataType: "text",
		Nullable: false,
		Comment:  "Hashed password for authentication",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "salt",
		DataType: "text",
		Nullable: false,
		Comment:  "Password salt",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "last_login",
		DataType: "datetime",
		Nullable: true,
		Comment:  "Last login timestamp",
	})
	userTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	// Create groups table for group membership
	groupTable := schema.NewTableSchema("groups")
	groupTable.AddField(schema.ColumnData{
		Name:       "id",
		DataType:   "text",
		Nullable:   false,
		PrimaryKey: true,
		Comment:    "Group ID",
	})
	groupTable.AddField(schema.ColumnData{
		Name:     "group_dn",
		DataType: "text",
		Nullable: false,
		Unique:   true,
		Comment:  "Group DN",
	})
	groupTable.AddField(schema.ColumnData{
		Name:     "member_dn",
		DataType: "text",
		Nullable: false,
		Index:    true,
		Comment:  "Member DN",
	})
	groupTable.AddField(schema.ColumnData{
		Name:     "created_at",
		DataType: "datetime",
		Nullable: false,
		Comment:  "Creation timestamp",
	})

	sch.AddTable(entryTable)
	sch.AddTable(userTable)
	sch.AddTable(groupTable)
	return sch
}