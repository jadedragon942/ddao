# RESTD - RESTful Microservice with DDAO

A RESTful microservice example built with the [Huma framework](https://huma.rocks/) demonstrating pluggable storage engines using DDAO (Dynamic Database Access Object).

## Features

- **RESTful API**: Built with Huma framework for modern Go API development
- **Pluggable Storage**: Support for multiple database backends through DDAO
- **OpenAPI Documentation**: Automatic API documentation generation
- **Storage Engines**: SQLite, PostgreSQL, CockroachDB, YugabyteDB, and TiDB support
- **CLI Interface**: Built-in CLI with configuration options
- **Health Checks**: Monitoring endpoint for service health

## Supported Storage Engines

| Engine | Connection String Example | Description |
|--------|---------------------------|-------------|
| SQLite | `restd.db` or `:memory:` | Local file or in-memory database |
| PostgreSQL | `postgres://user:pass@localhost:5432/db?sslmode=disable` | PostgreSQL database |
| CockroachDB | `postgres://root@localhost:26257/defaultdb?sslmode=disable` | CockroachDB cluster |
| YugabyteDB | `postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable` | YugabyteDB cluster |
| TiDB | `root:password@tcp(localhost:4000)/test` | TiDB cluster |

## Quick Start

### 1. Install Dependencies

```bash
cd examples/restd
go mod tidy
```

### 2. Run with SQLite (Default)

```bash
go run . --port 8080 --storage sqlite --connection "restd.db"
```

### 3. Run with PostgreSQL

```bash
go run . --port 8080 --storage postgres --connection "postgres://user:pass@localhost:5432/restd?sslmode=disable"
```

### 4. Access the API

- **API Documentation**: http://localhost:8080/docs
- **OpenAPI Spec**: http://localhost:8080/openapi.json
- **Health Check**: http://localhost:8080/health

## API Endpoints

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Create a new user |
| GET | `/users/{userId}` | Get user by ID |
| GET | `/users?email={email}` | Get user by email |
| PUT | `/users/{userId}` | Update user |
| DELETE | `/users/{userId}` | Delete user |

### Posts

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/posts` | Create a new post |
| GET | `/posts/{postId}` | Get post by ID |
| PUT | `/posts/{postId}` | Update post |
| DELETE | `/posts/{postId}` | Delete post |

## Example Usage

### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe",
    "profile": {"location": "San Francisco", "bio": "Software Engineer"}
  }'
```

### Get a User

```bash
curl http://localhost:8080/users/user_1234567890
```

### Create a Post

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_1234567890",
    "title": "My First Post",
    "content": "This is the content of my first post.",
    "published": true,
    "metadata": {"category": "technology", "tags": ["go", "api"]}
  }'
```

## CLI Options

```bash
Usage: restd [options]

Options:
  -p, --port int             Port to listen on (default: 8080)
  -s, --storage string       Storage engine (sqlite, postgres, cockroach, yugabyte, tidb) (default: sqlite)
  -c, --connection string    Database connection URL (default: restd.db)
  -d, --database string      Database name (default: restd)
  -h, --help                 Show help
```

## Data Models

### User

```json
{
  "id": "user_1234567890",
  "email": "john@example.com",
  "name": "John Doe",
  "profile": {
    "location": "San Francisco",
    "bio": "Software Engineer"
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Post

```json
{
  "id": "post_1234567890",
  "user_id": "user_1234567890",
  "title": "My First Post",
  "content": "This is the content of my first post.",
  "metadata": {
    "category": "technology",
    "tags": ["go", "api"]
  },
  "published": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

## Architecture

The application demonstrates several key architectural patterns:

### Pluggable Storage Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Huma     │────│   RestD     │────│    DDAO     │
│ REST API    │    │  Service    │    │     ORM     │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                    ┌────────────────────────┼────────────────────────┐
                    │                        │                        │
              ┌─────────────┐          ┌─────────────┐          ┌─────────────┐
              │   SQLite    │          │ PostgreSQL  │          │  TiDB/etc   │
              │   Storage   │          │   Storage   │          │   Storage   │
              └─────────────┘          └─────────────┘          └─────────────┘
```

### Key Components

1. **Storage Factory**: Creates appropriate storage engines based on configuration
2. **RestD Service**: Business logic layer handling API operations
3. **DDAO Integration**: Database operations through the unified DDAO interface
4. **Huma Framework**: HTTP routing, validation, and OpenAPI documentation

## Development

### Running Tests

```bash
go test ./...
```

### Adding New Endpoints

1. Define input/output models in `models.go`
2. Add handler methods to `RestdService` in `handlers.go`
3. Register routes in `main.go`

### Adding New Storage Engines

1. Implement the `storage.Storage` interface
2. Add the engine to `StorageEngine` enum in `config.go`
3. Update the `StorageFactory.CreateStorage` method

## Configuration Examples

### Production PostgreSQL

```bash
go run . \
  --port 8080 \
  --storage postgres \
  --connection "postgres://api_user:secure_password@prod-db:5432/restd_prod?sslmode=require"
```

### Development with CockroachDB

```bash
go run . \
  --port 8080 \
  --storage cockroach \
  --connection "postgres://root@localhost:26257/restd_dev?sslmode=disable"
```

### In-Memory Testing

```bash
go run . \
  --port 8080 \
  --storage sqlite \
  --connection ":memory:"
```

## Contributing

This example demonstrates best practices for building RESTful microservices with Go. Contributions and improvements are welcome!

## License

This example is part of the DDAO project and follows the same MIT license.