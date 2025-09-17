# LDAP Server Example Makefile

.PHONY: build run test clean help deps

# Default target
all: build

# Build the LDAP server
build:
	@echo "Building LDAP server..."
	go build -o ldapd .

# Run the server with default settings
run: build
	@echo "Starting LDAP server on port 1389..."
	./ldapd -port 1389 -v

# Run with custom database
run-postgres: build
	@echo "Starting LDAP server with PostgreSQL backend..."
	./ldapd -port 1389 -db "postgres://ldap_user:ldap_pass@localhost/ldapdb?sslmode=disable" -v

# Run tests
test:
	@echo "Running unit tests..."
	go test -short ./...

# Run integration tests (requires running server)
test-integration:
	@echo "Running integration tests..."
	go test -v ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f ldapd
	rm -f ldap.db
	rm -f *.log

# Run with different databases for testing
run-sqlite: build
	@echo "Starting LDAP server with SQLite backend..."
	./ldapd -port 1389 -db "ldap.db" -v

run-memory: build
	@echo "Starting LDAP server with in-memory SQLite..."
	./ldapd -port 1389 -db ":memory:" -v

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go install -v .

# Testing with different configurations
test-config:
	@echo "Testing with different configurations..."
	./ldapd -port 1389 -base-dn "dc=test,dc=local" -bind-dn "cn=root,dc=test,dc=local" &
	sleep 2
	pkill ldapd

# Docker build (if Dockerfile exists)
docker-build:
	@echo "Building Docker image..."
	docker build -t ddao-ldapd .

docker-run:
	@echo "Running in Docker..."
	docker run -p 1389:389 ddao-ldapd

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the LDAP server binary"
	@echo "  run             - Run the server with default settings (port 1389)"
	@echo "  run-postgres    - Run with PostgreSQL backend"
	@echo "  run-sqlite      - Run with SQLite backend"
	@echo "  run-memory      - Run with in-memory SQLite"
	@echo "  test            - Run unit tests"
	@echo "  test-integration- Run integration tests"
	@echo "  bench           - Run benchmarks"
	@echo "  deps            - Install dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  dev-setup       - Setup development environment"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run in Docker container"
	@echo "  help            - Show this help"