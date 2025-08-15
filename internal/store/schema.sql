-- Enhanced users table with additional fields
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),
    bio TEXT,
    password_hash TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);

-- Sessions table for SCS
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

-- Index for session expiry cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expiry);