.PHONY: run build test clean dev deps sqlite postgres cockroach yugabyte tidb

# Default target
run: deps
	go run .

# Build the application
build: deps
	go build -o restd .

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f restd
	go clean

# Development with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Run with different storage engines
sqlite:
	go run . --storage sqlite --connection "restd.db" --port 8080

sqlite-memory:
	go run . --storage sqlite --connection ":memory:" --port 8080

postgres:
	go run . --storage postgres --connection "postgres://postgres:password@localhost:5432/restd?sslmode=disable" --port 8080

cockroach:
	go run . --storage cockroach --connection "postgres://root@localhost:26257/defaultdb?sslmode=disable" --port 8080

yugabyte:
	go run . --storage yugabyte --connection "postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable" --port 8080

tidb:
	go run . --storage tidb --connection "root:@tcp(localhost:4000)/test" --port 8080

# Docker development environments
docker-postgres:
	docker run --name restd-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=restd -p 5432:5432 -d postgres:13

docker-cockroach:
	docker run -d --name restd-cockroach -p 26257:26257 cockroachdb/cockroach:latest start-single-node --insecure

docker-yugabyte:
	docker run -d --name restd-yugabyte -p 5433:5433 yugabytedb/yugabyte:latest bin/yugabyted start --daemon=false

docker-tidb:
	docker run --name restd-tidb -d -p 4000:4000 pingcap/tidb:latest

# Stop and remove Docker containers
docker-clean:
	docker stop restd-postgres restd-cockroach restd-yugabyte restd-tidb 2>/dev/null || true
	docker rm restd-postgres restd-cockroach restd-yugabyte restd-tidb 2>/dev/null || true

# Test API endpoints (requires curl)
test-api: test-health test-users test-posts

test-health:
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/health | jq .

test-users:
	@echo "Testing user endpoints..."
	@echo "Creating user..."
	@curl -s -X POST http://localhost:8080/users \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","name":"Test User","profile":{"role":"tester"}}' | jq .
	@echo "Getting user by email..."
	@curl -s "http://localhost:8080/users?email=test@example.com" | jq .

test-posts:
	@echo "Testing post endpoints..."
	@echo "Note: Run test-users first to create a user"

# Show API documentation URLs
show-docs:
	@echo "API Documentation: http://localhost:8080/docs"
	@echo "OpenAPI Spec: http://localhost:8080/openapi.json"
	@echo "Health Check: http://localhost:8080/health"

# Help target
help:
	@echo "Available targets:"
	@echo "  run           - Run the application with default settings"
	@echo "  build         - Build the application binary"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  dev           - Run with auto-reload (requires air)"
	@echo "  deps          - Install dependencies"
	@echo ""
	@echo "Storage engines:"
	@echo "  sqlite        - Run with SQLite storage"
	@echo "  sqlite-memory - Run with in-memory SQLite"
	@echo "  postgres      - Run with PostgreSQL storage"
	@echo "  cockroach     - Run with CockroachDB storage"
	@echo "  yugabyte      - Run with YugabyteDB storage"
	@echo "  tidb          - Run with TiDB storage"
	@echo ""
	@echo "Docker:"
	@echo "  docker-*      - Start Docker containers for databases"
	@echo "  docker-clean  - Stop and remove Docker containers"
	@echo ""
	@echo "Testing:"
	@echo "  test-api      - Test API endpoints"
	@echo "  test-health   - Test health endpoint"
	@echo "  test-users    - Test user endpoints"
	@echo "  test-posts    - Test post endpoints"
	@echo ""
	@echo "Documentation:"
	@echo "  show-docs     - Show API documentation URLs"