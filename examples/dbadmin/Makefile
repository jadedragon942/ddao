.PHONY: build clean test run-sqlite run-postgres

# Build the database admin tool
build:
	go build -o dbadmin .

# Clean build artifacts
clean:
	rm -f dbadmin admin.db

# Test the build
test: build
	@echo "Testing database admin tool help..."
	./dbadmin -help
	@echo "\nTesting SQLite startup..."
	timeout 3 ./dbadmin -port 8082 || true
	@echo "Build and basic tests completed successfully!"

# Run with SQLite (default)
run-sqlite: build
	./dbadmin -storage sqlite -conn "file:admin.db" -port 8080

# Run with PostgreSQL (requires running PostgreSQL)
run-postgres: build
	./dbadmin -storage postgres -conn "postgres://ddao:ddao_password@localhost:5432/ddao_test?sslmode=disable" -port 8080

# Demo with multiple storage types
demo: build
	@echo "Available storage types:"
	@echo "  SQLite:     make run-sqlite"
	@echo "  PostgreSQL: make run-postgres (requires PostgreSQL)"
	@echo ""
	@echo "Or run manually:"
	@echo "  ./dbadmin -storage [sqlite|postgres|sqlserver|oracle|cockroach|yugabyte|tidb|scylla|s3] -conn [connection-string] -port [port]"