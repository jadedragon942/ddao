sqlite:
	# SQLite (works standalone)
	docker-compose -f docker-compose.sqlite.yml run --rm ddao-sqlite

postgres:
	# PostgreSQL (with database)
	docker-compose -f docker-compose.postgres.yml up ddao-postgres

tidb:
	# MySQL/TiDB (with database)
	docker-compose -f docker-compose.mysql.yml up ddao-mysql

roach:
	# CockroachDB (with database)
	docker-compose -f docker-compose.cockroach.yml up ddao-cockroach

yuga:
	# YugabyteDB (with database)
	docker-compose -f docker-compose.yugabyte.yml up ddao-yugabyte
