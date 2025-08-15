-- +goose Up
-- Add password hash field to users table
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- Create sessions table for SCS
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

-- Index for session expiry cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expiry);

-- +goose Down
-- Rollback sessions and password changes
DROP INDEX IF EXISTS idx_sessions_expiry;
DROP TABLE IF EXISTS sessions;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;