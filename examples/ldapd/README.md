# LDAP Server Example with DDAO

This example demonstrates how to build a simple LDAP (Lightweight Directory Access Protocol) server using DDAO as the backend storage. The server provides basic LDAP operations including authentication, search, add, modify, and delete operations.

## Features

- **LDAP Protocol Support**: Basic LDAP operations (bind, search, add, modify, delete)
- **Multi-Database Backend**: Uses DDAO for storage, supporting SQLite, PostgreSQL, MySQL, and more
- **Authentication**: User authentication with hashed passwords
- **Hierarchical Directory**: Supports DN-based hierarchical directory structure
- **Group Management**: Basic group membership functionality
- **TLS Support**: Optional TLS encryption support

## Architecture

The LDAP server uses three main DDAO tables:

1. **entries**: Stores LDAP directory entries (DN, object class, attributes)
2. **users**: Stores user authentication data (password hashes, salts)
3. **groups**: Stores group membership relationships

## Quick Start

### 1. Build and Run

```bash
cd examples/ldapd
go mod tidy
go build
./ldapd
```

### 2. Command Line Options

```bash
./ldapd -help
```

Available options:
- `-port`: LDAP server port (default: 389)
- `-db`: Database file path (default: "ldap.db")
- `-bind-dn`: Admin bind DN (default: "cn=admin,dc=example,dc=com")
- `-bind-pw`: Admin password (default: "admin")
- `-base-dn`: Base DN (default: "dc=example,dc=com")
- `-v`: Verbose logging

### 3. Example Usage

Start the server:
```bash
./ldapd -port 1389 -v
```

## Testing the LDAP Server

The server implements a simplified text-based protocol for testing. You can connect using telnet:

```bash
telnet localhost 1389
```

### Basic Commands

1. **Authenticate as admin**:
   ```
   BIND cn=admin,dc=example,dc=com admin
   ```

2. **Search for entries**:
   ```
   SEARCH dc=example,dc=com
   ```

3. **Add a new user**:
   ```
   ADD cn=john,ou=users,dc=example,dc=com objectClass=person givenName=John sn=Doe userPassword=secret
   ```

4. **Search for the new user**:
   ```
   SEARCH cn=john,ou=users,dc=example,dc=com
   ```

5. **Delete a user**:
   ```
   DELETE cn=john,ou=users,dc=example,dc=com
   ```

6. **Get help**:
   ```
   HELP
   ```

## Initial Directory Structure

The server automatically creates the following initial structure:

```
dc=example,dc=com                     (Base DN - organization)
├── cn=admin,dc=example,dc=com        (Admin user)
├── ou=users,dc=example,dc=com        (Users organizational unit)
└── ou=groups,dc=example,dc=com       (Groups organizational unit)
```

## Using with Standard LDAP Clients

While this example uses a simplified protocol for demonstration, you can extend it to support the full LDAP protocol using libraries like:

- [go-ldap](https://github.com/go-ldap/ldap) for Go LDAP client/server
- [goldap](https://github.com/lor00x/goldap) for LDAP message parsing

## Configuration with Different Databases

### PostgreSQL Backend

```bash
# Modify main.go to use PostgreSQL instead of SQLite
import "github.com/jadedragon942/ddao/storage/postgres"

storage := postgres.New()
err := storage.Connect(ctx, "postgres://user:password@localhost/ldapdb?sslmode=disable")
```

### MySQL Backend

```bash
# Modify main.go to use MySQL instead of SQLite
import "github.com/jadedragon942/ddao/storage/mysql"

storage := mysql.New()
err := storage.Connect(ctx, "user:password@tcp(localhost:3306)/ldapdb")
```

## Security Considerations

This example is for demonstration purposes. For production use, consider:

1. **TLS/SSL**: Enable TLS encryption for secure connections
2. **Password Security**: Use stronger password hashing (bcrypt, scrypt)
3. **Access Control**: Implement proper LDAP access control lists (ACLs)
4. **Input Validation**: Add comprehensive input validation and sanitization
5. **Rate Limiting**: Implement connection and request rate limiting
6. **Logging**: Add comprehensive audit logging

## Advanced Features

### TLS Support

To enable TLS, you'll need certificates:

```bash
# Generate self-signed certificates for testing
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

### Schema Validation

Extend the server to support LDAP schema validation by:

1. Defining object classes and attribute types
2. Validating entries against schema rules
3. Supporting schema queries

### Replication

For high availability, consider implementing:

1. Master-slave replication
2. Multi-master replication
3. Distributed consensus (using Raft)

## Performance Tuning

- **Indexing**: Add database indexes for frequently searched attributes
- **Connection Pooling**: Use database connection pooling for high load
- **Caching**: Implement entry caching for frequently accessed data
- **Batching**: Support batch operations for bulk updates

## Monitoring and Metrics

Consider adding:

- Connection count metrics
- Operation latency metrics
- Error rate monitoring
- Database performance metrics

## Testing

```bash
# Run basic functionality tests
go test ./...

# Test with different databases
DB_TYPE=postgres go test ./...
DB_TYPE=mysql go test ./...
```

## License

This example is part of the DDAO project and follows the same MIT license.