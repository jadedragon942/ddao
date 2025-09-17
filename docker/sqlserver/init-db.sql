-- SQL Server initialization script for DDAO testing
-- Create test databases

-- Create ddao_test database
IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'ddao_test')
BEGIN
    CREATE DATABASE ddao_test;
    PRINT 'Database ddao_test created successfully.';
END
ELSE
    PRINT 'Database ddao_test already exists.';

-- Create ddao_dev database
IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'ddao_dev')
BEGIN
    CREATE DATABASE ddao_dev;
    PRINT 'Database ddao_dev created successfully.';
END
ELSE
    PRINT 'Database ddao_dev already exists.';

-- Switch to ddao_test database and create sample schema
USE ddao_test;

-- Create a test table to verify everything works
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='test_table' AND xtype='U')
BEGIN
    CREATE TABLE test_table (
        id NVARCHAR(255) PRIMARY KEY,
        name NVARCHAR(255) NOT NULL,
        email NVARCHAR(255) UNIQUE,
        profile NVARCHAR(MAX),
        active BIT DEFAULT 1,
        score FLOAT DEFAULT 0.0,
        created_at DATETIME2 DEFAULT GETDATE()
    );
    PRINT 'Test table created successfully.';
END;

-- Insert sample data
IF NOT EXISTS (SELECT * FROM test_table WHERE id = 'test-user-1')
BEGIN
    INSERT INTO test_table (id, name, email, profile, active, score)
    VALUES
        ('test-user-1', 'John Doe', 'john@example.com', '{"role": "admin", "preferences": {"theme": "dark"}}', 1, 95.5),
        ('test-user-2', 'Jane Smith', 'jane@example.com', '{"role": "user", "preferences": {"theme": "light"}}', 1, 87.2),
        ('test-user-3', 'Bob Wilson', 'bob@example.com', '{"role": "moderator", "preferences": {"theme": "auto"}}', 0, 92.8);
    PRINT 'Sample data inserted successfully.';
END;

-- Create ddao_dev schema
USE ddao_dev;
PRINT 'Switched to ddao_dev database for development use.';

-- Verify setup
SELECT
    'SQL Server Setup Complete' as Status,
    @@VERSION as Version,
    GETDATE() as Timestamp;

PRINT 'SQL Server initialization script completed successfully!';