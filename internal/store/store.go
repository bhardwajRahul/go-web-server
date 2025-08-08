// Package store provides database access and query execution functionality.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries.
type Store struct {
	*Queries // Embed sqlc-generated queries

	db *pgxpool.Pool
}

// PoolConfig holds database connection pool configuration.
type PoolConfig struct {
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// NewStore creates a new store instance with database connection pool.
func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	// Default pool configuration
	poolConfig := PoolConfig{
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: 0,
		MaxConnIdleTime: 0,
	}

	return NewStoreWithConfig(ctx, databaseURL, poolConfig)
}

// NewStoreWithConfig creates a new store instance with custom pool configuration.
func NewStoreWithConfig(ctx context.Context, databaseURL string, poolConfig PoolConfig) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set connection pool settings from config
	config.MaxConns = poolConfig.MaxConns
	config.MinConns = poolConfig.MinConns
	config.MaxConnLifetime = poolConfig.MaxConnLifetime
	config.MaxConnIdleTime = poolConfig.MaxConnIdleTime

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{
		db:      db,
		Queries: New(db),
	}, nil
}

// NewStoreWithDB creates a new store instance with an existing database pool.
func NewStoreWithDB(db *pgxpool.Pool) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.db.Close()
}

// DB returns the underlying database connection pool for advanced operations.
func (s *Store) DB() *pgxpool.Pool {
	return s.db
}

// BeginTx starts a new transaction.
func (s *Store) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.db.Begin(ctx)
}

// WithTx returns a new Store that will execute queries within the given transaction.
func (s *Store) WithTx(tx pgx.Tx) *Store {
	return &Store{
		db:      s.db,
		Queries: s.Queries.WithTx(tx),
	}
}

// InitSchema initializes the database schema using the schema.sql file.
// This is kept here for compatibility, but migrations are preferred.
func (s *Store) InitSchema(ctx context.Context) error {
	schema := `
		-- Enhanced users table with additional fields
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
	`

	_, err := s.db.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
