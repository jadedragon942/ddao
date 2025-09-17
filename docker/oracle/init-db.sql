-- Oracle initialization script for DDAO testing
-- This script runs as SYS user

-- Create tablespaces for DDAO
CREATE TABLESPACE ddao_data
DATAFILE '/opt/oracle/oradata/XE/ddao_data.dbf'
SIZE 100M
AUTOEXTEND ON
NEXT 10M
MAXSIZE UNLIMITED;

CREATE TABLESPACE ddao_temp
TEMPFILE '/opt/oracle/oradata/XE/ddao_temp.dbf'
SIZE 50M
AUTOEXTEND ON
NEXT 5M
MAXSIZE UNLIMITED;

-- Create DDAO user
CREATE USER ddao_user IDENTIFIED BY ddao_password
DEFAULT TABLESPACE ddao_data
TEMPORARY TABLESPACE ddao_temp
QUOTA UNLIMITED ON ddao_data;

-- Grant necessary privileges
GRANT CREATE SESSION TO ddao_user;
GRANT CREATE TABLE TO ddao_user;
GRANT CREATE SEQUENCE TO ddao_user;
GRANT CREATE VIEW TO ddao_user;
GRANT CREATE PROCEDURE TO ddao_user;
GRANT CREATE TRIGGER TO ddao_user;
GRANT ALTER ANY TABLE TO ddao_user;
GRANT DROP ANY TABLE TO ddao_user;

-- Additional privileges for testing
GRANT CREATE ANY INDEX TO ddao_user;
GRANT DROP ANY INDEX TO ddao_user;

-- Show created user
SELECT username, default_tablespace, temporary_tablespace, created
FROM all_users
WHERE username = 'DDAO_USER';

-- Success message
SELECT 'Oracle DDAO user setup completed successfully!' as status FROM dual;