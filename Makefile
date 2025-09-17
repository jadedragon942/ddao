sqlite:
	# SQLite (works standalone)
	docker-compose -f docker-compose.sqlite.yml run --rm ddao-sqlite

postgres:
	# PostgreSQL (with database)
	docker-compose -f docker-compose.postgres.yml up ddao-postgres

tidb:
	# MySQL/TiDB (with database)
	docker-compose -f docker-compose.mysql.yml up ddao-mysql

cockroachdb:
	# CockroachDB (with database)
	docker-compose -f docker-compose.cockroach.yml up ddao-cockroach

yugabyte:
	# YugabyteDB (with database)
	docker-compose -f docker-compose.yugabyte.yml up ddao-yugabyte

scylla:
	# ScyllaDB (with database)
	docker-compose -f docker-compose.scylla.yml up ddao-scylla

oracle:
	# Oracle
	docker-compose -f docker-compose.oracle.yml up ddao-oracle

sqlserver:
	# SQLServer
	docker-compose -f docker-compose.sqlserver.yml up ddao-sqlserver

clean:
	docker-compose down --remove-orphans --volumes
	docker system prune -f
