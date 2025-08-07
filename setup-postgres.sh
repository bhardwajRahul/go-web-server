#!/bin/bash

# Setup PostgreSQL for go-web-server development
# This script creates the necessary user and database

echo "Setting up PostgreSQL for go-web-server..."

# Create user and database
psql -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';"
psql -U postgres -c "CREATE DATABASE gowebserver OWNER \"user\";"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE gowebserver TO \"user\";"

echo "PostgreSQL setup complete!"
echo "Database: gowebserver"
echo "User: user"
echo "Password: password"
echo ""
echo "Connection string: postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@localhost:5432/gowebserver?sslmode=disable"