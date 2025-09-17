# Oracle Development Container for DDAO
FROM gvenzl/oracle-xe:21-slim

# Set environment variables
ENV ORACLE_PASSWORD=OraclePassword123
ENV ORACLE_DATABASE=DDAO
ENV APP_USER=ddao_user
ENV APP_USER_PASSWORD=ddao_password

# Copy initialization scripts
COPY init-db.sql /container-entrypoint-initdb.d/01-init-db.sql
COPY setup-user.sql /container-entrypoint-initdb.d/02-setup-user.sql

# Expose Oracle port
EXPOSE 1521

# The base image already handles startup, so we don't need a custom CMD