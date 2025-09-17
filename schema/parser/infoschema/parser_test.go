package infoschema

import (
	"database/sql"
	"testing"

	"github.com/jadedragon942/ddao/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT 1
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			user_id INTEGER NOT NULL,
			published_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE INDEX idx_posts_user_id ON posts(user_id);
		CREATE INDEX idx_posts_published ON posts(published_at);
		CREATE UNIQUE INDEX idx_users_email ON users(email);
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestNewParser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	parser := NewParser(db)
	assert.NotNil(t, parser)
	assert.Equal(t, db, parser.db)
}

func TestParseSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	adapter := NewSQLiteAdapter(db)

	schema, err := adapter.ParseSchema("main")
	require.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "main", schema.DatabaseName)

	assert.Contains(t, schema.Tables, "users")
	assert.Contains(t, schema.Tables, "posts")

	usersTable := schema.Tables["users"]
	assert.Equal(t, "users", usersTable.TableName)
	assert.Contains(t, usersTable.Fields, "id")
	assert.Contains(t, usersTable.Fields, "username")
	assert.Contains(t, usersTable.Fields, "email")

	idField := usersTable.Fields["id"]
	assert.True(t, idField.PrimaryKey)
	assert.True(t, idField.AutoIncrement)
	assert.False(t, idField.Nullable)

	usernameField := usersTable.Fields["username"]
	assert.True(t, usernameField.Unique)
	assert.False(t, usernameField.Nullable)

	postsTable := schema.Tables["posts"]
	assert.Equal(t, "posts", postsTable.TableName)
	assert.Contains(t, postsTable.Fields, "id")
	assert.Contains(t, postsTable.Fields, "title")
	assert.Contains(t, postsTable.Fields, "user_id")
}

func TestParseTableFromName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	adapter := NewSQLiteAdapter(db)

	tableSchema, err := adapter.ParseTableFromName("main", "users")
	require.NoError(t, err)
	assert.NotNil(t, tableSchema)
	assert.Equal(t, "users", tableSchema.TableName)

	assert.Contains(t, tableSchema.Fields, "id")
	assert.Contains(t, tableSchema.Fields, "username")
	assert.Contains(t, tableSchema.Fields, "email")

	idField := tableSchema.Fields["id"]
	assert.True(t, idField.PrimaryKey)
	assert.True(t, idField.AutoIncrement)

	nonExistentTable, err := adapter.ParseTableFromName("main", "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, nonExistentTable)
}

func TestGetTableNames(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	adapter := NewSQLiteAdapter(db)

	tableNames, err := adapter.GetTableNames("main")
	require.NoError(t, err)
	assert.Contains(t, tableNames, "users")
	assert.Contains(t, tableNames, "posts")
	assert.Len(t, tableNames, 2)
}

func TestConvertColumn(t *testing.T) {
	parser := &Parser{}

	column := InfoSchemaColumn{
		ColumnName:    "test_col",
		DataType:      "VARCHAR",
		IsNullable:    "NO",
		ColumnKey:     "PRI",
		Extra:         "auto_increment",
		ColumnComment: "Test column",
	}

	columnData := parser.convertColumn(column)

	assert.Equal(t, "test_col", columnData.Name)
	assert.Equal(t, "VARCHAR", columnData.DataType)
	assert.False(t, columnData.Nullable)
	assert.True(t, columnData.PrimaryKey)
	assert.True(t, columnData.AutoIncrement)
	assert.True(t, columnData.Index)
	assert.Equal(t, "Test column", columnData.Comment)
}

func TestConvertColumnUnique(t *testing.T) {
	parser := &Parser{}

	column := InfoSchemaColumn{
		ColumnName: "unique_col",
		DataType:   "VARCHAR",
		IsNullable: "YES",
		ColumnKey:  "UNI",
	}

	columnData := parser.convertColumn(column)

	assert.Equal(t, "unique_col", columnData.Name)
	assert.True(t, columnData.Nullable)
	assert.True(t, columnData.Unique)
	assert.True(t, columnData.Index)
	assert.False(t, columnData.PrimaryKey)
	assert.False(t, columnData.AutoIncrement)
}

func TestConvertColumnWithDefault(t *testing.T) {
	parser := &Parser{}

	defaultValue := "default_value"
	column := InfoSchemaColumn{
		ColumnName:    "col_with_default",
		DataType:      "VARCHAR",
		IsNullable:    "YES",
		ColumnDefault: &defaultValue,
	}

	columnData := parser.convertColumn(column)

	assert.Equal(t, "col_with_default", columnData.Name)
	assert.Equal(t, "default_value", columnData.Default)
}

func TestProcessIndexes(t *testing.T) {
	parser := &Parser{}
	tableSchema := &schema.TableSchema{
		TableName:  "test_table",
		Fields:     make(map[string]schema.ColumnData),
		Indexes:    []string{},
		UniqueKeys: []string{},
	}

	indexes := []InfoSchemaIndex{
		{
			IndexName:  "idx_test",
			ColumnName: "col1",
			NonUnique:  1,
		},
		{
			IndexName:  "idx_unique",
			ColumnName: "col2",
			NonUnique:  0,
		},
		{
			IndexName:  "PRIMARY",
			ColumnName: "id",
			NonUnique:  0,
		},
	}

	parser.processIndexes(tableSchema, indexes)

	assert.Contains(t, tableSchema.Indexes, "idx_test")
	assert.Contains(t, tableSchema.Indexes, "idx_unique")
	assert.NotContains(t, tableSchema.Indexes, "PRIMARY")

	assert.Contains(t, tableSchema.UniqueKeys, "idx_unique")
	assert.NotContains(t, tableSchema.UniqueKeys, "idx_test")
}
