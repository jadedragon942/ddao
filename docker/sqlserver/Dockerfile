# SQL Server Development Container for DDAO
FROM mcr.microsoft.com/mssql/server:2022-latest

# Set environment variables
ENV ACCEPT_EULA=Y
ENV SA_PASSWORD=YourStrong@Passw0rd123
ENV MSSQL_PID=Developer

# Create directory for initialization scripts
USER root
RUN mkdir -p /opt/mssql-tools/init
COPY init-db.sql /opt/mssql-tools/init/

# Create startup script
COPY startup.sh /opt/mssql-tools/
RUN chmod +x /opt/mssql-tools/startup.sh

# Switch back to mssql user
USER mssql

# Expose SQL Server port
EXPOSE 1433

# Use custom startup script
CMD ["/opt/mssql-tools/startup.sh"]