# Database Support for DynDAO

This document outlines the database support added to the DynDAO project.

## Supported Databases

The following databases are now supported with full CRUD operations:

### 1. SQLite ✅
- **Package**: `github.com/jadedragon942/ddao/storage/sqlite`
- **Driver**: `github.com/mattn/go-sqlite3`
- **Status**: Fully implemented and tested
- **Connection**: In-memory (`:memory:`) for tests, file-based for production
- **Features**:
  - UPSERT using `INSERT OR REPLACE`
  - JSON support
  - Full CRUD operations

### 2. PostgreSQL ✅
- **Package**: `github.com/jadedragon942/ddao/storage/postgres`
- **Driver**: `github.com/lib/pq`
- **Status**: Fully implemented
- **Connection**: Standard PostgreSQL connection string
- **Features**:
  - UPSERT using `INSERT ... ON CONFLICT DO UPDATE`
  - JSONB support for better JSON performance
  - Full CRUD operations

### 3. CockroachDB ✅
- **Package**: `github.com/jadedragon942/ddao/storage/cockroach`
- **Driver**: `github.com/lib/pq` (PostgreSQL-compatible)
- **Status**: Fully implemented
- **Connection**: CockroachDB connection string
- **Features**:
  - Uses CockroachDB-specific data types (STRING, INT8, etc.)
  - UPSERT using CockroachDB's native `UPSERT` statement
  - JSONB support
  - Full CRUD operations

### 4. YugabyteDB ✅
- **Package**: `github.com/jadedragon942/ddao/storage/yugabyte`
- **Driver**: `github.com/jackc/pgx/v5/stdlib` (PostgreSQL-compatible)
- **Status**: Fully implemented
- **Connection**: YugabyteDB connection string
- **Features**:
  - PostgreSQL-compatible syntax
  - UPSERT using `INSERT ... ON CONFLICT DO UPDATE`
  - JSONB support for distributed JSON operations
  - Full CRUD operations

### 5. TiDB ✅
- **Package**: `github.com/jadedragon942/ddao/storage/tidb`
- **Driver**: `github.com/go-sql-driver/mysql` (MySQL-compatible)
- **Status**: Fully implemented
- **Connection**: MySQL-compatible connection string
- **Features**:
  - MySQL-compatible syntax with TiDB optimizations
  - UPSERT using MySQL's `REPLACE INTO`
  - JSON support for modern JSON operations
  - Full CRUD operations

## Common Interface

All database implementations conform to the `storage.Storage` interface:

```go
type Storage interface {
    Connect(ctx context.Context, connStr string) error
    CreateTables(ctx context.Context, schema *schema.Schema) error
    Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error)
    Update(ctx context.Context, obj *object.Object) (bool, error)
    FindByID(ctx context.Context, tblName, id string) (*object.Object, error)
    FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error)
    DeleteByID(ctx context.Context, tblName, id string) (bool, error)
    ResetConnection(ctx context.Context) error
}
```

## Testing

### Comprehensive Test Suite

Each database implementation includes:

1. **Basic Storage Test**: Using `storagetest.StorageTest()`
2. **Comprehensive CRUD Test**: Using `storagetest.CRUDTest()`
3. **Local Integration Tests**: For testing against local database instances

### Test Coverage

The test suite covers:
- ✅ **CREATE**: Insert new objects with validation
- ✅ **READ**: Find objects by ID and by arbitrary fields
- ✅ **UPDATE**: Modify existing objects
- ✅ **DELETE**: Remove objects and verify deletion
- ✅ **UPSERT**: Insert or update objects atomically
- ✅ **Edge Cases**: Non-existent objects, error handling
- ✅ **Data Types**: Text, JSON, integers, nullable fields

### Running Tests

```bash
# Run SQLite tests (always work)
go test github.com/jadedragon942/ddao/storage/sqlite

# Run tests with local database (requires setup)
go test github.com/jadedragon942/ddao/storage/postgres -run TestPostgreSQLLocal

# Run all tests
go test github.com/jadedragon942/ddao/...
```

## Usage Examples

### SQLite
```go
storage := sqlite.New()
err := storage.Connect(ctx, "file.db") // or ":memory:"
```

### PostgreSQL
```go
storage := postgres.New()
err := storage.Connect(ctx, "postgres://user:pass@host:5432/db?sslmode=disable")
```

### CockroachDB
```go
storage := cockroach.New()
err := storage.Connect(ctx, "postgres://root@host:26257/db?sslmode=disable")
```

### YugabyteDB
```go
storage := yugabyte.New()
err := storage.Connect(ctx, "postgres://yugabyte@host:5433/yugabyte?sslmode=disable")
```

### TiDB
```go
storage := tidb.New()
err := storage.Connect(ctx, "user:password@tcp(host:4000)/db")
```

## Local Development Setup

For testing with local database instances:

### PostgreSQL
```bash
docker run --name postgres-test -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=testdb -p 5432:5432 -d postgres:13
```

### CockroachDB
```bash
docker run -d --name=roach --hostname=roach -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start-single-node --insecure
```

### YugabyteDB
```bash
docker run -d --name yugabyte -p 7000:7000 -p 9000:9000 -p 5433:5433 -p 9042:9042 yugabytedb/yugabyte:latest bin/yugabyted start --daemon=false
```

### TiDB
```bash
docker run --name tidb-server -d -p 4000:4000 pingcap/tidb:latest
```

## Implementation Notes

### Data Type Mapping

Each database implementation includes intelligent data type mapping:
- `TEXT/VARCHAR/CHAR` → Database-specific text types
- `INTEGER/INT` → Database-specific integer types
- `JSON` → Native JSON/JSONB where available
- `DATETIME/TIMESTAMP` → Database-specific date/time types
- `BOOLEAN` → Native boolean types
- `BLOB` → Binary data types

### UPSERT Strategy

- **SQLite**: `INSERT OR REPLACE`
- **PostgreSQL/YugabyteDB**: `INSERT ... ON CONFLICT DO UPDATE`
- **CockroachDB**: Native `UPSERT`
- **TiDB**: `REPLACE INTO`

### Consistency

All implementations ensure:
- Consistent field ordering for reliable data mapping
- Proper handling of nullable vs non-nullable fields
- Atomic UPSERT operations
- Error handling for connection and query failures

## Performance Considerations

- **SQLite**: Best for single-user applications, embedded systems
- **PostgreSQL**: Excellent for complex queries, ACID compliance
- **CockroachDB**: Best for distributed, globally consistent applications
- **YugabyteDB**: Optimal for distributed SQL with high availability
- **TiDB**: Great for horizontal scaling with MySQL compatibility

This implementation provides a solid foundation for multi-database support in DynDAO while maintaining a consistent interface across all database types.