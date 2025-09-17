# S3 Storage Demo for DDAO

This example demonstrates how to use DDAO with Amazon S3 (or S3-compatible storage) as the backend storage system.

## Features

- **S3 Storage Backend**: Uses Amazon S3 or S3-compatible storage (like MinIO) for data persistence
- **Object Storage**: Stores DDAO objects as JSON files in S3 with hierarchical key structure
- **Schema Management**: Automatically creates table metadata in S3
- **CRUD Operations**: Full support for Create, Read, Update, Delete operations
- **Search Capabilities**: Supports both ID-based and key-based object searches

## Architecture

The S3 storage adapter stores data with the following structure:

```
s3://bucket/prefix/
├── _schema.json                           # Schema metadata
├── tables/
│   ├── users/
│   │   ├── _metadata.json                 # Table metadata
│   │   └── objects/
│   │       ├── user123.json               # User objects
│   │       └── user456.json
│   └── documents/
│       ├── _metadata.json
│       └── objects/
│           ├── doc1.json                  # Document objects
│           └── doc2.json
```

## Quick Start

### Prerequisites

You'll need either:
1. **AWS S3**: Valid AWS credentials and an S3 bucket
2. **MinIO**: Local MinIO server for development/testing

### Setting up MinIO (Recommended for Development)

```bash
# Install MinIO (macOS)
brew install minio/stable/minio

# Or download from https://min.io/download

# Start MinIO server
mkdir -p ~/minio-data
minio server ~/minio-data

# Default access: http://localhost:9000
# Default credentials: minioadmin / minioadmin
```

Create a bucket named `ddao-demo-bucket` using the MinIO web interface or CLI.

### Running the Demo

```bash
cd examples/s3demo
go mod tidy
go build

# For MinIO (default)
./s3demo -v

# For AWS S3
./s3demo -conn "s3://your-bucket/ddao-data?region=us-east-1" -v

# For custom S3-compatible endpoint
./s3demo -conn "s3://bucket/prefix?region=us-east-1&endpoint=https://s3.example.com" -v
```

### Connection String Format

The S3 connection string uses the following format:

```
s3://bucket/prefix?region=region&endpoint=endpoint&verbose=true
```

Parameters:
- `bucket`: S3 bucket name (required)
- `prefix`: Key prefix for all objects (optional)
- `region`: AWS region (default: us-east-1)
- `endpoint`: Custom S3 endpoint for S3-compatible services (optional)
- `verbose`: Enable verbose logging (optional)

### Examples

**AWS S3:**
```
s3://my-production-bucket/ddao-data?region=us-west-2
```

**MinIO:**
```
s3://ddao-demo-bucket/test-data?region=us-east-1&endpoint=http://localhost:9000
```

**DigitalOcean Spaces:**
```
s3://my-space/ddao?region=nyc3&endpoint=https://nyc3.digitaloceanspaces.com
```

**Wasabi:**
```
s3://my-bucket/ddao?region=us-east-1&endpoint=https://s3.wasabisys.com
```

## Demo Operations

The demo performs the following operations:

1. **Schema Creation**: Creates tables for users and documents
2. **User Creation**: Creates a user with profile data
3. **Document Creation**: Creates multiple documents owned by the user
4. **Search Operations**: Demonstrates ID-based and key-based searches
5. **Update Operations**: Updates user profile information
6. **Cleanup**: Optionally removes created objects

## Configuration

### AWS Credentials

For AWS S3, configure credentials using one of:

1. **Environment Variables:**
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

2. **AWS Credentials File** (`~/.aws/credentials`):
   ```ini
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   ```

3. **IAM Roles** (when running on EC2)

### MinIO Credentials

For MinIO, set the credentials as environment variables:

```bash
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
```

## Production Considerations

### Performance

- **Latency**: S3 operations have higher latency than local databases
- **Throughput**: Optimize for batch operations when possible
- **Costs**: Be aware of request costs, especially for list operations

### Best Practices

1. **Bucket Organization**: Use meaningful prefixes to organize data
2. **Access Patterns**: Design your key structure for efficient searches
3. **Monitoring**: Use CloudWatch/monitoring for S3 operations
4. **Backup**: Consider versioning and cross-region replication
5. **Security**: Use IAM policies and bucket policies for access control

### Limitations

- **Transactions**: S3 doesn't support ACID transactions
- **Consistency**: S3 is eventually consistent for some operations
- **Search**: Limited search capabilities compared to databases
- **Concurrency**: No built-in locking mechanisms

## Integration Examples

### Web Application

```go
func createS3Storage(bucket, prefix, region string) storage.Storage {
    storage := s3.New()
    connStr := fmt.Sprintf("s3://%s/%s?region=%s", bucket, prefix, region)

    err := storage.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatal(err)
    }

    return storage
}
```

### Microservice

```go
func initStorage() *orm.ORM {
    bucket := os.Getenv("S3_BUCKET")
    region := os.Getenv("AWS_REGION")

    storage := createS3Storage(bucket, "microservice-data", region)
    schema := createSchema()

    return orm.New(schema).WithStorage(storage)
}
```

### Development vs Production

```go
func createStorage(env string) storage.Storage {
    switch env {
    case "development":
        // Use MinIO for development
        return createS3Storage("dev-bucket", "dev", "us-east-1")
    case "production":
        // Use AWS S3 for production
        return createS3Storage("prod-bucket", "prod", "us-west-2")
    default:
        // Use SQLite for local testing
        storage := sqlite.New()
        storage.Connect(context.Background(), ":memory:")
        return storage
    }
}
```

## Troubleshooting

### Common Issues

1. **Access Denied**: Check bucket permissions and AWS credentials
2. **Bucket Not Found**: Verify bucket name and region
3. **Connection Timeout**: Check network connectivity and endpoint
4. **Invalid Credentials**: Verify AWS credentials are correct

### Debug Mode

Enable verbose logging to see detailed S3 operations:

```bash
./s3demo -conn "s3://bucket/prefix?verbose=true" -v
```

### Testing Connectivity

```go
// Test S3 connectivity
func testS3Connection(connStr string) error {
    storage := s3.New()
    err := storage.Connect(context.Background(), connStr)
    if err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }
    defer storage.ResetConnection(context.Background())

    // Try to create a simple schema
    schema := schema.New()
    schema.SetDatabaseName("test")

    return storage.CreateTables(context.Background(), schema)
}
```

## License

This example is part of the DDAO project and follows the same MIT license.