# DDAO - Dynamic Database Access Object

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)](LICENSE)

DDAO is a flexible, multi-database ORM (Object-Relational Mapping) library for Go that provides a unified interface for working with different database systems. It supports dynamic schema definition and offers consistent CRUD operations across multiple database backends.

## 🚀 Features

- **Multi-Database Support**: SQLite, PostgreSQL, CockroachDB, YugabyteDB, TiDB, and ScyllaDB
- **Dynamic Schema Definition**: Define table schemas programmatically with flexible field types
- **Unified Interface**: Same API works across all supported databases
- **Type-Safe Operations**: Built-in type conversion and validation
- **UPSERT Support**: Atomic insert-or-update operations
- **JSON Support**: Native JSON field handling for modern applications
- **Transaction Support**: Full transaction support with transaction-aware CRUD operations
- **Comprehensive Testing**: Full test suite with CRUD operation coverage
- **Connection Management**: Proper connection lifecycle management

## 📦 Supported Databases

| Database | Driver | Status | Features |
|----------|--------|--------|----------|
| **SQLite** | `github.com/mattn/go-sqlite3` | ✅ Full Support | In-memory & file-based, UPSERT |
| **PostgreSQL** | `github.com/lib/pq` | ✅ Full Support | JSONB, Advanced SQL features |
| **CockroachDB** | `github.com/lib/pq` | ✅ Full Support | Distributed, Native UPSERT |
| **YugabyteDB** | `github.com/jackc/pgx/v5` | ✅ Full Support | Distributed SQL, PostgreSQL compatible |
| **TiDB** | `github.com/go-sql-driver/mysql` | ✅ Full Support | Horizontal scaling, MySQL compatible |
| **ScyllaDB** | `github.com/scylladb/gocql` | ✅ Full Support | High-performance NoSQL, Cassandra compatible |

## 🛠 Installation

```bash
go get github.com/jadedragon942/ddao
```

### Database-Specific Dependencies

The core library includes all necessary drivers. For specific databases:

```bash
# SQLite (included)
go get github.com/mattn/go-sqlite3

# PostgreSQL (included)
go get github.com/lib/pq

# YugabyteDB (included)
go get github.com/jackc/pgx/v5

# MySQL/TiDB (included)
go get github.com/go-sql-driver/mysql

# ScyllaDB (included)
go get github.com/scylladb/gocql
```

## 🏗 Architecture

DDAO consists of four main components:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    ORM      │────│   Object    │────│   Schema    │────│   Storage   │
│             │    │             │    │             │    │             │
│ High-level  │    │ Data model  │    │ Table def   │    │ DB backends │
│ operations  │    │ with fields │    │ with types  │    │ (5 drivers) │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Core Components

1. **Object**: Represents a data record with dynamic fields
2. **Schema**: Defines table structure and field types
3. **Storage**: Database-specific implementations (SQLite, PostgreSQL, etc.)
4. **ORM**: High-level interface that orchestrates the components

## 📚 Quick Start

### 1. Define Your Schema

```go
package main

import (
    "context"
    "log"

    "github.com/jadedragon942/ddao/object"
    "github.com/jadedragon942/ddao/orm"
    "github.com/jadedragon942/ddao/schema"
    "github.com/jadedragon942/ddao/storage/sqlite"
)

func createUserSchema() *schema.Schema {
    sch := schema.New()
    sch.SetDatabaseName("myapp")

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
        Unique:   true,
        Comment:  "User email address",
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
```

### 2. Initialize ORM with Database

```go
func main() {
    ctx := context.Background()

    // Create schema
    schema := createUserSchema()

    // Initialize storage backend (SQLite in this example)
    storage := sqlite.New()
    err := storage.Connect(ctx, "users.db") // or ":memory:" for in-memory
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer storage.ResetConnection(ctx)

    // Create ORM instance
    orm := orm.New(schema).WithStorage(storage)

    // Create tables
    err = storage.CreateTables(ctx, schema)
    if err != nil {
        log.Fatal("Failed to create tables:", err)
    }

    // Now you can use the ORM for operations
    performCRUDOperations(ctx, orm)
}
```

### 3. Perform CRUD Operations

```go
func performCRUDOperations(ctx context.Context, orm *orm.ORM) {
    // CREATE - Insert a new user
    user := object.New()
    user.TableName = "users"
    user.ID = "user123"
    user.Fields = map[string]any{
        "email":      "john@example.com",
        "profile":    `{"name": "John Doe", "age": 30}`,
        "created_at": "2024-01-01T00:00:00Z",
    }

    data, created, err := orm.Insert(ctx, user)
    if err != nil {
        log.Printf("Insert failed: %v", err)
        return
    }
    log.Printf("User created: %t, Data: %s", created, string(data))

    // READ - Find user by ID
    foundUser, err := orm.FindByID(ctx, "users", "user123")
    if err != nil {
        log.Printf("Find failed: %v", err)
        return
    }
    if foundUser != nil {
        email, _ := foundUser.GetString("email")
        log.Printf("Found user: %s, Email: %s", foundUser.ID, email)
    }

    // READ - Find user by email
    userByEmail, err := orm.FindByKey(ctx, "users", "email", "john@example.com")
    if err != nil {
        log.Printf("Find by email failed: %v", err)
        return
    }
    if userByEmail != nil {
        log.Printf("Found user by email: %s", userByEmail.ID)
    }

    // UPDATE - Modify user profile
    user.Fields["profile"] = `{"name": "John Doe", "age": 31, "city": "New York"}`
    updated, err := orm.Storage.Update(ctx, user)
    if err != nil {
        log.Printf("Update failed: %v", err)
        return
    }
    log.Printf("User updated: %t", updated)

    // DELETE - Remove user
    deleted, err := orm.DeleteByID(ctx, "users", "user123")
    if err != nil {
        log.Printf("Delete failed: %v", err)
        return
    }
    log.Printf("User deleted: %t", deleted)
}
```

## 🗄 Database-Specific Usage

### SQLite

```go
import "github.com/jadedragon942/ddao/storage/sqlite"

storage := sqlite.New()
err := storage.Connect(ctx, "database.db")        // File-based
// or
err := storage.Connect(ctx, ":memory:")           // In-memory
```

### PostgreSQL

```go
import "github.com/jadedragon942/ddao/storage/postgres"

storage := postgres.New()
err := storage.Connect(ctx, "postgres://user:password@localhost:5432/dbname?sslmode=disable")
```

### CockroachDB

```go
import "github.com/jadedragon942/ddao/storage/cockroach"

storage := cockroach.New()
err := storage.Connect(ctx, "postgres://root@localhost:26257/defaultdb?sslmode=disable")
```

### YugabyteDB

```go
import "github.com/jadedragon942/ddao/storage/yugabyte"

storage := yugabyte.New()
err := storage.Connect(ctx, "postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable")
```

### TiDB

```go
import "github.com/jadedragon942/ddao/storage/tidb"

storage := tidb.New()
err := storage.Connect(ctx, "root:password@tcp(localhost:4000)/test")
```

### ScyllaDB

```go
import "github.com/jadedragon942/ddao/storage/scylla"

storage := scylla.New()
err := storage.Connect(ctx, "localhost:9042,192.168.1.2:9042/mykeyspace?consistency=quorum&timeout=10s")
```

## 🔧 Advanced Features

### Working with Objects

The `object.Object` type provides convenient methods for type-safe field access:

```go
user := object.New()
user.TableName = "users"
user.ID = "user123"
user.SetField("email", "john@example.com")
user.SetField("age", 30)
user.SetField("active", true)

// Type-safe getters
email, exists := user.GetString("email")      // Returns: "john@example.com", true
age, exists := user.GetInt64("age")           // Returns: 30, true
active, exists := user.GetBool("active")      // Returns: true, true
missing, exists := user.GetString("missing")  // Returns: "", false

// Generic getter
value, exists := user.GetField("email")       // Returns: "john@example.com", true
```

### Schema Definition with Advanced Options

```go
table := schema.NewTableSchema("products")

// Primary key with auto-increment
table.AddField(schema.ColumnData{
    Name:          "id",
    DataType:      "integer",
    Nullable:      false,
    PrimaryKey:    true,
    AutoIncrement: true,
})

// Indexed field with default value
table.AddField(schema.ColumnData{
    Name:     "category",
    DataType: "text",
    Nullable: false,
    Default:  "general",
    Index:    true,
})

// JSON field for flexible data
table.AddField(schema.ColumnData{
    Name:     "attributes",
    DataType: "json",
    Nullable: true,
    Comment:  "Product attributes in JSON format",
})

// Unique constraint
table.AddField(schema.ColumnData{
    Name:     "sku",
    DataType: "text",
    Nullable: false,
    Unique:   true,
})
```

### Error Handling

```go
// Connection errors
storage := postgres.New()
err := storage.Connect(ctx, "invalid-connection-string")
if err != nil {
    log.Printf("Connection failed: %v", err)
    // Handle connection error
}

// Operation errors
obj, err := orm.FindByID(ctx, "users", "nonexistent")
if err != nil {
    log.Printf("Query failed: %v", err)
    // Handle query error
}
if obj == nil {
    log.Println("User not found")
    // Handle not found case
}

// Validation errors during insert/update
user := object.New()
user.TableName = "users"
// Missing required fields...
_, _, err = orm.Insert(ctx, user)
if err != nil {
    log.Printf("Validation failed: %v", err)
    // Handle validation error
}
```

## 🧪 Testing

DDAO includes a comprehensive test suite. Run tests for specific databases:

```bash
# Run SQLite tests (always work)
go test ./storage/sqlite/

# Run all tests (requires database setup for some)
go test ./...

# Run with verbose output
go test -v ./storage/sqlite/
```

### Running Tests with Local Databases

For testing against actual database instances:

```bash
# PostgreSQL
docker run --name postgres-test -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=testdb -p 5432:5432 -d postgres:13
go test ./storage/postgres/ -run TestPostgreSQLLocal

# CockroachDB
docker run -d --name=roach -p 26257:26257 cockroachdb/cockroach:latest start-single-node --insecure
go test ./storage/cockroach/ -run TestCockroachDBLocal

# YugabyteDB
docker run -d --name yugabyte -p 5433:5433 yugabytedb/yugabyte:latest bin/yugabyted start --daemon=false
go test ./storage/yugabyte/ -run TestYugabyteDBLocal

# TiDB
docker run --name tidb-server -d -p 4000:4000 pingcap/tidb:latest
go test ./storage/tidb/ -run TestTiDBLocal
```

### Custom Tests

You can use the `storagetest` package to test your own storage implementations:

```go
import "github.com/jadedragon942/ddao/storagetest"

func TestMyStorage(t *testing.T) {
    storage := MyCustomStorage()
    // ... setup storage ...

    // Run comprehensive CRUD tests
    storagetest.CRUDTest(t, storage)

    // Run basic storage tests
    storagetest.StorageTest(t, storage)
}
```

## 🏗 Implementation Details

### Storage Interface

All database implementations conform to the `storage.Storage` interface:

```go
type Storage interface {
    Connect(ctx context.Context, connStr string) error
    CreateTables(ctx context.Context, schema *Schema) error
    Insert(ctx context.Context, obj *Object) ([]byte, bool, error)
    Update(ctx context.Context, obj *Object) (bool, error)
    FindByID(ctx context.Context, tblName, id string) (*Object, error)
    FindByKey(ctx context.Context, tblName, key, value string) (*Object, error)
    DeleteByID(ctx context.Context, tblName, id string) (bool, error)
    ResetConnection(ctx context.Context) error

    // Transaction support
    BeginTx(ctx context.Context) (*sql.Tx, error)
    CommitTx(tx *sql.Tx) error
    RollbackTx(tx *sql.Tx) error
    InsertTx(ctx context.Context, tx *sql.Tx, obj *Object) ([]byte, bool, error)
    UpdateTx(ctx context.Context, tx *sql.Tx, obj *Object) (bool, error)
    FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*Object, error)
    FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*Object, error)
    DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error)
}
```

### UPSERT Behavior

DDAO implements database-specific UPSERT (insert-or-update) operations:

- **SQLite**: `INSERT OR REPLACE INTO ...`
- **PostgreSQL/YugabyteDB**: `INSERT ... ON CONFLICT DO UPDATE SET ...`
- **CockroachDB**: `UPSERT INTO ...`
- **TiDB**: `REPLACE INTO ...`

### Data Type Mapping

DDAO automatically maps generic types to database-specific types:

| Generic Type | SQLite | PostgreSQL | CockroachDB | YugabyteDB | TiDB |
|--------------|--------|------------|-------------|------------|------|
| `text` | TEXT | TEXT | STRING | TEXT | TEXT |
| `integer` | INTEGER | INTEGER | INT8 | INTEGER | BIGINT |
| `json` | TEXT | JSONB | JSONB | JSONB | JSON |
| `datetime` | TEXT | TIMESTAMP | TIMESTAMPTZ | TIMESTAMP | TIMESTAMP |
| `boolean` | INTEGER | BOOLEAN | BOOL | BOOLEAN | BOOLEAN |

### Field Ordering Consistency

DDAO ensures consistent field ordering across operations to prevent data corruption during read/write operations. This is particularly important when dealing with dynamic schemas where field order can vary.

## 💳 Transaction Support

DDAO provides full transaction support with transaction-aware versions of all CRUD operations. This allows you to perform multiple operations atomically within a single database transaction.

### Transaction Methods

All storage implementations provide these transaction methods:

- `BeginTx(ctx)` - Start a new transaction
- `CommitTx(tx)` - Commit the transaction
- `RollbackTx(tx)` - Rollback the transaction
- `InsertTx(ctx, tx, obj)` - Insert within a transaction
- `UpdateTx(ctx, tx, obj)` - Update within a transaction
- `FindByIDTx(ctx, tx, tblName, id)` - Find by ID within a transaction
- `FindByKeyTx(ctx, tx, tblName, key, value)` - Find by key within a transaction
- `DeleteByIDTx(ctx, tx, tblName, id)` - Delete by ID within a transaction

### Transaction Usage Example

```go
func performTransactionalOperations(ctx context.Context, orm *orm.ORM) error {
    // Begin a new transaction
    tx, err := orm.BeginTx(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Ensure transaction is properly handled
    defer func() {
        if err != nil {
            orm.RollbackTx(tx)
        }
    }()

    // Create first user within transaction
    user1 := object.New()
    user1.TableName = "users"
    user1.ID = "user1"
    user1.Fields = map[string]any{
        "email":      "user1@example.com",
        "created_at": "2024-01-01T00:00:00Z",
    }

    _, created, err := orm.InsertTx(ctx, tx, user1)
    if err != nil {
        return fmt.Errorf("failed to insert user1: %w", err)
    }
    if !created {
        return fmt.Errorf("user1 was not created")
    }

    // Create second user within the same transaction
    user2 := object.New()
    user2.TableName = "users"
    user2.ID = "user2"
    user2.Fields = map[string]any{
        "email":      "user2@example.com",
        "created_at": "2024-01-01T00:00:00Z",
    }

    _, created, err = orm.InsertTx(ctx, tx, user2)
    if err != nil {
        return fmt.Errorf("failed to insert user2: %w", err)
    }

    // Verify both users exist within the transaction
    foundUser1, err := orm.FindByIDTx(ctx, tx, "users", "user1")
    if err != nil {
        return fmt.Errorf("failed to find user1: %w", err)
    }
    if foundUser1 == nil {
        return fmt.Errorf("user1 not found in transaction")
    }

    foundUser2, err := orm.FindByIDTx(ctx, tx, "users", "user2")
    if err != nil {
        return fmt.Errorf("failed to find user2: %w", err)
    }
    if foundUser2 == nil {
        return fmt.Errorf("user2 not found in transaction")
    }

    // Update user1 within transaction
    foundUser1.Fields["email"] = "updated_user1@example.com"
    updated, err := orm.UpdateTx(ctx, tx, foundUser1)
    if err != nil {
        return fmt.Errorf("failed to update user1: %w", err)
    }
    if !updated {
        return fmt.Errorf("user1 was not updated")
    }

    // Commit the transaction - all operations succeed or all fail
    err = orm.CommitTx(tx)
    if err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    log.Println("Transaction completed successfully")
    return nil
}
```

### Transaction Best Practices

1. **Always handle errors properly**: Use defer to ensure transactions are rolled back on errors
2. **Keep transactions short**: Long-running transactions can cause lock contention
3. **Use appropriate isolation levels**: The default isolation level is usually sufficient
4. **Test transaction boundaries**: Ensure your business logic properly handles transaction rollbacks
5. **Handle deadlocks**: Be prepared to retry transactions that fail due to deadlocks

### Error Handling with Transactions

```go
func safeTransactionalOperation(ctx context.Context, orm *orm.ORM) error {
    tx, err := orm.BeginTx(ctx)
    if err != nil {
        return err
    }

    // Use a flag to track if we should commit or rollback
    shouldCommit := false
    defer func() {
        if shouldCommit {
            if commitErr := orm.CommitTx(tx); commitErr != nil {
                log.Printf("Failed to commit transaction: %v", commitErr)
            }
        } else {
            if rollbackErr := orm.RollbackTx(tx); rollbackErr != nil {
                log.Printf("Failed to rollback transaction: %v", rollbackErr)
            }
        }
    }()

    // Perform your operations...
    user := object.New()
    user.TableName = "users"
    user.ID = "safe_user"
    user.Fields = map[string]any{"email": "safe@example.com"}

    _, _, err = orm.InsertTx(ctx, tx, user)
    if err != nil {
        return err // Will trigger rollback via defer
    }

    // If we reach here, mark for commit
    shouldCommit = true
    return nil
}
```

## 🔒 Best Practices

### 1. Schema Design
- Use descriptive field names and comments
- Define appropriate indexes for frequently queried fields
- Use JSON fields sparingly and only for truly dynamic data
- Set proper nullable constraints

### 2. Connection Management
- Always defer `ResetConnection()` after connecting
- Use connection pooling for high-throughput applications
- Handle connection errors gracefully

### 3. Error Handling
- Check for `nil` objects when using Find operations
- Validate required fields before Insert/Update operations
- Use appropriate error logging for debugging

### 4. Performance
- Use appropriate indexes for your query patterns
- Consider using `FindByKey` with indexed fields instead of scanning
- For high-volume operations, consider batch processing

### 5. Security
- Use parameterized queries (handled automatically by DDAO)
- Validate input data before database operations
- Use appropriate database user permissions

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Write tests** for your changes
4. **Ensure all tests pass**: `go test ./...`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Setup

```bash
# Clone the repository
git clone https://github.com/jadedragon942/ddao.git
cd ddao

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: Check this README and the [DATABASE_SUPPORT.md](DATABASE_SUPPORT.md) file
- **Issues**: Report bugs and request features via [GitHub Issues](https://github.com/jadedragon942/ddao/issues)
- **Discussions**: Join the conversation in [GitHub Discussions](https://github.com/jadedragon942/ddao/discussions)

## 🙏 Acknowledgments

- Database driver maintainers for excellent Go database libraries
- The Go community for continuous innovation in database tooling
- Contributors who help improve DDAO

---

**DDAO** - Making multi-database development simple and consistent! 🚀# ddao
