-- +goose Up
-- Initial schema for users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),
    bio TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);

-- Insert sample data for development - The creators of Go
INSERT INTO users (email, name, bio) VALUES 
    ('robert@google.com', 'Robert Griesemer', 'Co-creator of Go programming language, designed at Google starting in 2007'),
    ('rob@google.com', 'Rob Pike', 'Co-creator of Go programming language, Unix pioneer and member of original Unix team'),
    ('ken@google.com', 'Ken Thompson', 'Co-creator of Go programming language, designed Unix and invented the B programming language')
ON CONFLICT(email) DO NOTHING;

-- +goose Down
-- Rollback initial schema
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;