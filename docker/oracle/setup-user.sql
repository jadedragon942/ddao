-- Oracle user setup script for DDAO testing
-- This script runs as ddao_user

-- Connect as ddao_user (this will be handled by the container)
-- CREATE TABLE to test functionality
CREATE TABLE test_table (
    id VARCHAR2(255) PRIMARY KEY,
    name VARCHAR2(255) NOT NULL,
    email VARCHAR2(255) UNIQUE,
    profile CLOB,
    active NUMBER(1) DEFAULT 1,
    score BINARY_DOUBLE DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO test_table (id, name, email, profile, active, score)
VALUES
    ('test-user-1', 'John Doe', 'john@example.com', '{"role": "admin", "preferences": {"theme": "dark"}}', 1, 95.5);

INSERT INTO test_table (id, name, email, profile, active, score)
VALUES
    ('test-user-2', 'Jane Smith', 'jane@example.com', '{"role": "user", "preferences": {"theme": "light"}}', 1, 87.2);

INSERT INTO test_table (id, name, email, profile, active, score)
VALUES
    ('test-user-3', 'Bob Wilson', 'bob@example.com', '{"role": "moderator", "preferences": {"theme": "auto"}}', 0, 92.8);

-- Commit the changes
COMMIT;

-- Create a sequence for testing
CREATE SEQUENCE test_seq
START WITH 1000
INCREMENT BY 1
CACHE 20;

-- Create a simple view for testing
CREATE VIEW active_users AS
SELECT id, name, email, score
FROM test_table
WHERE active = 1;

-- Verify setup
SELECT 'Oracle user schema setup completed successfully!' as status FROM dual;

-- Show table info
SELECT table_name, num_rows
FROM user_tables
WHERE table_name = 'TEST_TABLE';

-- Show sample data
SELECT COUNT(*) as sample_records
FROM test_table;