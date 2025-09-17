-- PostgreSQL initialization script for DDAO testing

-- Create test table
CREATE TABLE IF NOT EXISTS test_table (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    profile JSONB,
    active BOOLEAN DEFAULT true,
    score REAL DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO test_table (id, name, email, profile, active, score)
VALUES
    ('test-user-1', 'John Doe', 'john@example.com', '{"role": "admin", "preferences": {"theme": "dark"}}', true, 95.5),
    ('test-user-2', 'Jane Smith', 'jane@example.com', '{"role": "user", "preferences": {"theme": "light"}}', true, 87.2),
    ('test-user-3', 'Bob Wilson', 'bob@example.com', '{"role": "moderator", "preferences": {"theme": "auto"}}', false, 92.8)
ON CONFLICT (id) DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_test_table_email ON test_table(email);
CREATE INDEX IF NOT EXISTS idx_test_table_active ON test_table(active);

-- Verify setup
SELECT 'PostgreSQL setup completed successfully!' as status;