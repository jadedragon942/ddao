# DDAO Docker Setup

This directory contains Docker configurations for running DDAO with SQL Server, Oracle, and other databases for development and testing.

## 🐳 Quick Start

### Start All Databases

```bash
# Start all database services
docker-compose up -d

# Start specific databases
docker-compose up -d sqlserver oracle
```

### Run Tests

```bash
# Run tests against all databases
./docker/scripts/run-tests.sh --test

# Run tests against specific databases
./docker/scripts/run-tests.sh --services "sqlserver oracle" --test

# Run tests and cleanup afterward
./docker/scripts/run-tests.sh --test --cleanup
```

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## 📁 Directory Structure

```
docker/
├── sqlserver/
│   ├── Dockerfile           # SQL Server container
│   ├── startup.sh          # Custom startup script
│   └── init-db.sql         # Database initialization
├── oracle/
│   ├── Dockerfile          # Oracle container
│   ├── init-db.sql         # System-level setup
│   └── setup-user.sql      # User schema setup
└── scripts/
    ├── postgres-init.sql   # PostgreSQL initialization
    ├── mysql-init.sql      # MySQL initialization
    ├── test.Dockerfile     # Test runner container
    └── run-tests.sh        # Test automation script
```

## 🗄️ Database Services

### SQL Server
- **Container**: `ddao-sqlserver`
- **Port**: `1433`
- **SA Password**: `YourStrong@Passw0rd123`
- **Test Database**: `ddao_test`
- **Dev Database**: `ddao_dev`
- **Connection String**: `sqlserver://sa:YourStrong@Passw0rd123@localhost:1433?database=ddao_test`

### Oracle
- **Container**: `ddao-oracle`
- **Port**: `1521`
- **System Password**: `OraclePassword123`
- **Database**: `xe`
- **App User**: `ddao_user`
- **App Password**: `ddao_password`
- **Connection String**: `oracle://ddao_user:ddao_password@localhost:1521/xe`

### PostgreSQL
- **Container**: `ddao-postgres`
- **Port**: `5432`
- **Database**: `ddao_test`
- **User**: `ddao_user`
- **Password**: `ddao_password`
- **Connection String**: `postgres://ddao_user:ddao_password@localhost:5432/ddao_test?sslmode=disable`

### MySQL
- **Container**: `ddao-mysql`
- **Port**: `3306`
- **Database**: `ddao_test`
- **User**: `ddao_user`
- **Password**: `ddao_password`
- **Connection String**: `ddao_user:ddao_password@tcp(localhost:3306)/ddao_test`

## 🔧 Development Workflow

### 1. Start Development Environment

```bash
# Start all databases for development
docker-compose up -d

# Check service health
docker-compose ps
```

### 2. Run Specific Tests

```bash
# Test SQL Server only
SQLSERVER_TEST_URL="sqlserver://sa:YourStrong@Passw0rd123@localhost:1433?database=ddao_test" \
go test ./storage/sqlserver/ -v

# Test Oracle only
ORACLE_TEST_URL="oracle://ddao_user:ddao_password@localhost:1521/xe" \
go test ./storage/oracle/ -v
```

### 3. Full Test Suite

```bash
# Run all tests with proper environment variables
./docker/scripts/run-tests.sh --test
```

## 🚀 Advanced Usage

### Custom Databases

You can extend the setup by adding your own database configurations:

```yaml
# Add to docker-compose.yml
  custom-db:
    image: your-database:latest
    ports:
      - "5000:5000"
    environment:
      - DB_PASSWORD=your_password
    networks:
      - ddao-network
```

### Environment Variables

The test runner sets these environment variables automatically:

```bash
SQLSERVER_TEST_URL=sqlserver://sa:YourStrong@Passw0rd123@sqlserver:1433?database=ddao_test
ORACLE_TEST_URL=oracle://ddao_user:ddao_password@oracle:1521/xe
POSTGRES_TEST_URL=postgres://ddao_user:ddao_password@postgres:5432/ddao_test?sslmode=disable
MYSQL_TEST_URL=ddao_user:ddao_password@tcp(mysql:3306)/ddao_test
```

### Performance Tuning

For better performance during development:

```bash
# Allocate more memory to containers
docker-compose up -d --scale sqlserver=1 --scale oracle=1

# Use SSD volumes for better I/O
docker volume create --driver local \
  --opt type=none \
  --opt o=bind \
  --opt device=/path/to/ssd \
  ddao_ssd_data
```

## 🔍 Troubleshooting

### Common Issues

1. **Services not starting**: Check system resources and Docker daemon
2. **Connection refused**: Wait for health checks to pass
3. **Permission denied**: Ensure Docker daemon is running with proper permissions
4. **Port conflicts**: Check if ports 1433, 1521, 5432, 3306 are available

### Debugging

```bash
# Check service logs
docker-compose logs sqlserver
docker-compose logs oracle

# Inspect container health
docker inspect ddao-sqlserver | grep Health
docker inspect ddao-oracle | grep Health

# Connect to containers directly
docker exec -it ddao-sqlserver /bin/bash
docker exec -it ddao-oracle /bin/bash
```

### Reset Everything

```bash
# Stop and remove everything
docker-compose down -v --remove-orphans

# Remove all DDAO-related images
docker rmi $(docker images | grep ddao | awk '{print $3}')

# Rebuild everything
docker-compose build --no-cache
docker-compose up -d
```

## 📝 Notes

### SQL Server Notes
- Uses Developer Edition (free for development)
- Includes both `ddao_test` and `ddao_dev` databases
- Automatic database initialization on first startup
- Health checks ensure database is ready before tests

### Oracle Notes
- Uses Oracle XE 21c (free)
- Creates dedicated tablespace for DDAO
- Includes both system and user-level initialization
- May take longer to start up (2-3 minutes)

### Performance Notes
- First startup may be slow due to image downloads
- Oracle container requires significant memory (recommended: 2GB+)
- SQL Server Developer Edition is feature-complete but resource-intensive
- Consider using `--profile test` to avoid starting test runner by default