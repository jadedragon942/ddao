#!/bin/bash

# DDAO Test Runner Script
# This script helps run tests against all database backends using Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 DDAO Docker Test Runner${NC}"
echo "================================"

# Function to print status
print_status() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker and Docker Compose are available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH"
    exit 1
fi

# Parse command line arguments
SERVICES="all"
RUN_TESTS=false
CLEANUP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --services)
            SERVICES="$2"
            shift 2
            ;;
        --test)
            RUN_TESTS=true
            shift
            ;;
        --cleanup)
            CLEANUP=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --services SERVICES  Specify which services to start (all, sqlserver, oracle, postgres, mysql)"
            echo "  --test              Run tests after starting services"
            echo "  --cleanup           Stop and remove containers after tests"
            echo "  --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --services sqlserver --test"
            echo "  $0 --services all --test --cleanup"
            echo "  $0 --cleanup"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Function to start services
start_services() {
    local services=$1
    print_status "Starting $services services..."

    if [ "$services" = "all" ]; then
        docker-compose up -d sqlserver oracle postgres mysql
    else
        docker-compose up -d $services
    fi

    print_status "Waiting for services to be healthy..."

    # Wait for services to be healthy
    local max_wait=300  # 5 minutes
    local wait_time=0

    while [ $wait_time -lt $max_wait ]; do
        if [ "$services" = "all" ]; then
            if docker-compose ps | grep -q "healthy.*healthy.*healthy.*healthy"; then
                break
            fi
        else
            # For specific services, check if they're healthy
            if docker-compose ps $services | grep -q "healthy"; then
                break
            fi
        fi

        sleep 10
        wait_time=$((wait_time + 10))
        print_status "Waiting... (${wait_time}s/${max_wait}s)"
    done

    if [ $wait_time -ge $max_wait ]; then
        print_error "Services did not become healthy within $max_wait seconds"
        docker-compose logs
        exit 1
    fi

    print_success "Services are ready!"
}

# Function to run tests
run_tests() {
    print_status "Running DDAO tests..."

    if docker-compose --profile test run --rm ddao-test; then
        print_success "All tests passed!"
    else
        print_error "Some tests failed!"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up containers..."
    docker-compose down -v
    print_success "Cleanup completed!"
}

# Main execution
if [ "$CLEANUP" = true ] && [ "$RUN_TESTS" = false ]; then
    cleanup
    exit 0
fi

# Start services
start_services "$SERVICES"

# Run tests if requested
if [ "$RUN_TESTS" = true ]; then
    if run_tests; then
        print_success "Test run completed successfully!"
    else
        if [ "$CLEANUP" = true ]; then
            cleanup
        fi
        exit 1
    fi
fi

# Cleanup if requested
if [ "$CLEANUP" = true ]; then
    cleanup
else
    print_status "Services are running. Use 'docker-compose down' to stop them."
    print_status "Or run '$0 --cleanup' to stop and clean up."
fi

print_success "Script completed successfully!"