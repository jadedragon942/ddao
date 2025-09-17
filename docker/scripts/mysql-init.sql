-- MySQL initialization script for DDAO testing

-- Create test table
CREATE TABLE IF NOT EXISTS test_table (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    profile JSON,
    active BOOLEAN DEFAULT true,
    score DOUBLE DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT IGNORE INTO test_table (id, name, email, profile, active, score)
VALUES
    ('test-user-1', 'John Doe', 'john@example.com', '{"role": "admin", "preferences": {"theme": "dark"}}', true, 95.5),
    ('test-user-2', 'Jane Smith', 'jane@example.com', '{"role": "user", "preferences": {"theme": "light"}}', true, 87.2),
    ('test-user-3', 'Bob Wilson', 'bob@example.com', '{"role": "moderator", "preferences": {"theme": "auto"}}', false, 92.8);

-- Create indexes for better performance
CREATE INDEX idx_test_table_email ON test_table(email);
CREATE INDEX idx_test_table_active ON test_table(active);

-- Verify setup
SELECT 'MySQL setup completed successfully!' as status;