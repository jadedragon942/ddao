# Database Administration Tool

A simple web-based database administration tool that supports all DDAO storage adapters. This tool provides a user-friendly interface for connecting to databases, managing database schemas, and performing ALTER TABLE operations.

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

- **DSN Connection Screen**: Web-based interface for connecting to any storage adapter with interactive examples
- **Schema Management**: View database tables and their field definitions
- **ALTER TABLE Support**: Add new columns to existing tables through a simple web interface
- **Dynamic Connection**: Connect and disconnect from different storage adapters without restarting
- **Simple UI**: Clean, responsive HTML/CSS interface with minimal JavaScript for enhanced usability

## Usage

### Command Line Options

```bash
./dbadmin [options]
```

Options:
- `-port`: Port to run the web server on (default: 8080)

### Starting the Tool

```bash
# Build the tool
go build -o dbadmin .

# Start the web server
./dbadmin

# Or specify a custom port
./dbadmin -port 9000
```

The tool will start without any database connection. Navigate to the web interface to connect to your preferred storage adapter.

### Connecting to Storage Adapters

Once the web server is running, navigate to `http://localhost:8080` (or your specified port) to access the DSN connection screen. The interface provides:

1. **Storage Type Selection**: Choose from all supported storage adapters
2. **Connection String Input**: Dynamic placeholder and examples based on selected storage type
3. **Interactive Help**: Real-time connection string examples and format guidance

### Connection String Examples

The web interface provides interactive examples, but here are the formats for reference:

#### SQLite
```
file:database.db
:memory:
```

#### PostgreSQL
```
postgres://user:password@localhost:5432/dbname?sslmode=disable
```

#### SQL Server
```
sqlserver://user:password@localhost:1433?database=dbname
```

#### Oracle
```
oracle://user:password@localhost:1521/XE
```

#### CockroachDB
```
postgres://user:password@localhost:26257/dbname?sslmode=disable
```

#### YugabyteDB
```
postgres://user:password@localhost:5433/dbname?sslmode=disable
```

#### TiDB
```
user:password@tcp(localhost:4000)/dbname
```

#### ScyllaDB/Cassandra
```
localhost:9042/keyspace
```

#### Amazon S3
```
bucket-name
```

## Building

To build the database administration tool:

```bash
go build -o dbadmin .
```

## Web Interface

Once started, open your browser and navigate to `http://localhost:8080` (or the port you specified).

The web interface provides:

1. **Connection Screen**: Initial screen for connecting to any storage adapter with dynamic examples
2. **Dashboard**: Overview of the storage configuration and available tables (shown after connection)
3. **Tables**: Detailed view of all tables and their schemas
4. **Alter Table**: Interface for adding new columns to existing tables
5. **Disconnect**: Option to disconnect and switch to a different storage adapter

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