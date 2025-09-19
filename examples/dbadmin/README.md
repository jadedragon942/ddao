# Database Administration Tool

A simple web-based database administration tool that supports all DDAO storage adapters. This tool provides a user-friendly interface for managing database schemas and performing ALTER TABLE operations.

## Features

- **Multi-Storage Support**: Works with all DDAO storage adapters:
  - SQLite
  - PostgreSQL
  - SQL Server
  - Oracle
  - CockroachDB
  - YugabyteDB
  - TiDB
  - ScyllaDB/Cassandra
  - Amazon S3

- **Schema Management**: View database tables and their field definitions
- **ALTER TABLE Support**: Add new columns to existing tables through a simple web interface
- **Simple UI**: Clean, responsive HTML/CSS interface with no JavaScript dependencies

## Usage

### Command Line Options

```bash
./dbadmin [options]
```

Options:
- `-storage`: Storage type (default: sqlite)
- `-conn`: Connection string for the storage (default: file:admin.db)
- `-port`: Port to run the web server on (default: 8080)

### Examples

#### SQLite (Default)
```bash
go run . -storage sqlite -conn "file:test.db"
```

#### PostgreSQL
```bash
go run . -storage postgres -conn "postgres://user:password@localhost/dbname?sslmode=disable"
```

#### SQL Server
```bash
go run . -storage sqlserver -conn "sqlserver://user:password@localhost:1433?database=dbname"
```

#### Oracle
```bash
go run . -storage oracle -conn "oracle://user:password@localhost:1521/XE"
```

#### CockroachDB
```bash
go run . -storage cockroach -conn "postgres://user:password@localhost:26257/dbname?sslmode=disable"
```

#### YugabyteDB
```bash
go run . -storage yugabyte -conn "postgres://user:password@localhost:5433/dbname?sslmode=disable"
```

#### TiDB
```bash
go run . -storage tidb -conn "user:password@tcp(localhost:4000)/dbname"
```

#### ScyllaDB/Cassandra
```bash
go run . -storage scylla -conn "localhost:9042/keyspace"
```

#### Amazon S3
```bash
go run . -storage s3 -conn "bucket-name"
```

## Building

To build the database administration tool:

```bash
go build -o dbadmin .
```

## Web Interface

Once started, open your browser and navigate to `http://localhost:8080` (or the port you specified).

The web interface provides:

1. **Dashboard**: Overview of the storage configuration and available tables
2. **Tables**: Detailed view of all tables and their schemas
3. **Alter Table**: Interface for adding new columns to existing tables

## ALTER TABLE Operations

The tool supports adding new columns to existing tables. When you use the "Alter Table" feature:

1. Select the target table from the dropdown
2. Enter the new column name
3. Choose the data type (TEXT, INTEGER, REAL, BOOLEAN, DATETIME, JSON, BLOB)
4. Specify whether the column should allow NULL values
5. Click "Add Column" to execute the ALTER TABLE statement

The tool automatically maps generic data types to the appropriate storage-specific types.

## Limitations

- **S3 Storage**: ALTER TABLE operations are not supported since S3 is object storage
- **Data Viewing**: Currently shows table schemas only; data viewing functionality is planned for future versions
- **Column Operations**: Only supports adding columns; dropping or modifying columns is not supported
- **ScyllaDB/Cassandra**: Some data type mappings may vary based on the specific version

## Demo Data

The tool automatically creates sample tables (`users` and `products`) for demonstration purposes when it starts up. These tables can be used to test the ALTER TABLE functionality.