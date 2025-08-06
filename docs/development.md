# Development Guide

Local development setup and workflow for the Modern Go Stack.

## Quick Setup

### Prerequisites

- **Go 1.24+** - [Download](https://golang.org/dl/)
- **Git** - Version control

### Installation

```bash
# Clone and setup
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server

# Install tools and dependencies
go mod tidy
mage setup

# Generate code and start development server
mage dev
```

Server runs at `http://localhost:8080`

## Development Commands

### Daily Workflow

```bash
mage dev          # Start with hot reload
mage generate     # Generate SQLC + Templ code
mage fmt          # Format all code
mage quality      # Run vet + lint + vulncheck
mage clean        # Clean build artifacts
```

### Build Commands

```bash
mage build        # Production binary
mage run          # Build and run
mage ci           # Full CI pipeline
```

### Database Commands

```bash
mage migrate         # Run migrations up
mage migrateDown     # Rollback last migration
mage migrateStatus   # Show migration status
```

## Project Structure

```
go-web-server/
├── cmd/web/main.go           # Application entry point
├── internal/
│   ├── config/config.go      # Configuration management
│   ├── handler/              # HTTP request handlers
│   │   ├── home.go          # Home page handlers
│   │   ├── routes.go        # Route registration
│   │   └── user.go          # User CRUD handlers
│   ├── middleware/           # Custom middleware
│   │   ├── csrf.go          # CSRF protection
│   │   ├── errors.go        # Error handling
│   │   ├── sanitize.go      # Input sanitization
│   │   └── validation.go    # Request validation
│   ├── store/               # Database layer
│   │   ├── migrations/      # Goose migrations
│   │   ├── queries.sql     # SQL queries (source)
│   │   ├── queries.sql.go  # Generated Go code
│   │   └── store.go        # Store implementation
│   ├── ui/static/           # Static assets (embedded)
│   └── view/                # Templ templates
├── bin/                     # Compiled binaries
├── magefile.go             # Build automation
└── sqlc.yaml               # SQLC configuration
```

## Database Development

### Creating Migrations

```bash
# Create new migration
goose -dir internal/store/migrations create add_feature sql
```

Migration file structure:

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN avatar_url TEXT;
CREATE INDEX idx_users_avatar ON users(avatar_url);

-- +goose Down
DROP INDEX idx_users_avatar;
ALTER TABLE users DROP COLUMN avatar_url;
```

### Writing SQL Queries

Add to `internal/store/queries.sql`:

```sql
-- name: GetActiveUsers :many
SELECT * FROM users
WHERE is_active = 1
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email, name, bio, avatar_url)
VALUES (?, ?, ?, ?)
RETURNING *;
```

Generate Go code:

```bash
mage generate  # or sqlc generate
```

Use in handlers:

```go
func (h *UserHandler) ListActiveUsers(c echo.Context) error {
    ctx := c.Request().Context()
    users, err := h.store.GetActiveUsers(ctx)
    if err != nil {
        return handleError(err, c)
    }

    component := view.UserList(users)
    return component.Render(ctx, c.Response().Writer)
}
```

## Template Development

### Template Structure

Base layout (`internal/view/layout/base.templ`):

```go
package layout

templ Base(title string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <title>{title} - Go Web Server</title>
        <link rel="stylesheet" href="/static/css/pico.min.css">
    </head>
    <body>
        { children... }
        <script src="/static/js/htmx.min.js"></script>
    </body>
    </html>
}
```

Page template:

```go
package view

import "github.com/dunamismax/go-web-server/internal/view/layout"

templ Users() {
    @layout.Base("User Management") {
        <section>
            <h1>User Management</h1>
            @UserList()
        </section>
    }
}
```

### HTMX Integration

Form with CSRF:

```go
templ UserForm(user *store.User, token string) {
    <form hx-post="/users" hx-target="#user-list">
        <input type="hidden" name="csrf_token" value={token}/>
        <input type="text" name="name" required/>
        <input type="email" name="email" required/>
        <button type="submit">Create User</button>
    </form>
}
```

Dynamic updates:

```go
templ UserRow(user store.User) {
    <tr id={"user-" + strconv.FormatInt(user.ID, 10)}>
        <td>{user.Name}</td>
        <td>{user.Email}</td>
        <td>
            <button hx-delete={"/users/" + strconv.FormatInt(user.ID, 10)}
                    hx-confirm="Delete this user?"
                    hx-target={"#user-" + strconv.FormatInt(user.ID, 10)}
                    hx-swap="outerHTML">
                Delete
            </button>
        </td>
    </tr>
}
```

## Handler Development

### Handler Pattern

```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    ctx := c.Request().Context()

    // Input validation
    name := c.FormValue("name")
    email := c.FormValue("email")

    if name == "" || email == "" {
        return middleware.NewAppErrorWithDetails(
            middleware.ErrorTypeValidation,
            http.StatusBadRequest,
            "Validation failed",
            map[string]string{
                "name": "Name is required",
                "email": "Email is required",
            },
        ).WithContext(c)
    }

    // Database operation
    params := store.CreateUserParams{
        Name:  name,
        Email: email,
    }

    user, err := h.store.CreateUser(ctx, params)
    if err != nil {
        slog.Error("Failed to create user", "error", err)
        return middleware.NewAppError(
            middleware.ErrorTypeInternal,
            http.StatusInternalServerError,
            "Failed to create user",
        ).WithContext(c).WithInternal(err)
    }

    // Success response
    slog.Info("User created", "user_id", user.ID, "email", email)

    // Return updated list for HTMX
    users, err := h.store.ListUsers(ctx)
    if err != nil {
        return handleError(err, c)
    }

    component := view.UserList(users)
    return component.Render(ctx, c.Response().Writer)
}
```

### Error Handling

```go
// Database errors
user, err := h.store.GetUser(ctx, id)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return middleware.NewAppError(
            middleware.ErrorTypeNotFound,
            http.StatusNotFound,
            "User not found",
        ).WithContext(c)
    }

    return middleware.NewAppError(
        middleware.ErrorTypeInternal,
        http.StatusInternalServerError,
        "Database error occurred",
    ).WithContext(c).WithInternal(err)
}
```

## Configuration

### Development Environment

Create `.env` file:

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=localhost

# Database
DATABASE_URL=data/development.db
DATABASE_RUN_MIGRATIONS=true

# Application
APP_ENVIRONMENT=development
APP_DEBUG=true
APP_LOG_LEVEL=debug
APP_LOG_FORMAT=text

# Security (development)
SECURITY_ENABLE_CORS=true
SECURITY_ALLOWED_ORIGINS=*

# Features (development)
FEATURES_ENABLE_METRICS=true
FEATURES_ENABLE_PPROF=true
```

### Configuration File

`config.json` (optional):

```json
{
  "server": {
    "port": "8080",
    "read_timeout": "10s",
    "write_timeout": "10s"
  },
  "app": {
    "environment": "development",
    "debug": true,
    "log_level": "debug"
  },
  "database": {
    "url": "data/development.db",
    "run_migrations": true
  },
  "features": {
    "enable_metrics": true,
    "enable_pprof": true
  }
}
```

## Quality Checks

```bash
# Individual checks
mage fmt          # Format code with goimports
mage vet          # Static analysis
mage lint         # Comprehensive linting
mage vulncheck    # Security vulnerability scan

# Combined
mage quality      # All quality checks
mage ci           # Full CI pipeline
```

## Debugging

### Logging

```go
// Development - text format for readability
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Structured debugging
slog.Debug("Processing request",
    "method", c.Request().Method,
    "path", c.Request().URL.Path,
    "user_id", getUserID(c),
    "request_id", c.Response().Header().Get(echo.HeaderXRequestID))
```

### Database Debugging

```bash
# Connect to database
sqlite3 data/development.db

# Useful commands
.tables                          # List tables
.schema users                    # Show schema
.headers on                      # Show headers
.mode column                     # Column display
SELECT * FROM users LIMIT 5;    # Query data
```

### Metrics Monitoring

**Development Endpoints:**

```bash
# Application metrics (Prometheus format)
curl http://localhost:8080/metrics

# Health check with database connectivity
curl http://localhost:8080/health

# Enhanced health information
curl http://localhost:8080/health | jq .
```

**Key Metrics Available:**

- `http_requests_total` - HTTP request count by method, path, status
- `http_request_duration_seconds` - Request latency histograms
- `http_requests_in_flight` - Current active requests
- `database_connections_active` - Active database connections
- `database_queries_total` - Database query count by operation
- `htmx_requests_total` - HTMX-specific request tracking
- `csrf_tokens_generated_total` - CSRF token generation rate
- `users_created_total` - Business metric tracking

**Development Monitoring:**

```bash
# Simple monitoring setup with basic tools
watch -n 5 'curl -s http://localhost:8080/health | jq ".checks"'

# Monitor HTTP metrics
watch -n 2 'curl -s http://localhost:8080/metrics | grep http_requests_total'
```

## Adding New Features

1. **Plan database changes** - Create migration if needed
2. **Add SQL queries** - Write in `queries.sql`
3. **Generate code** - `mage generate`
4. **Create handlers** - Implement business logic
5. **Add routes** - Register in `routes.go`
6. **Create templates** - Build UI components
7. **Test** - `mage dev` and verify functionality

## Development Tools

### Air Configuration (`.air.toml`)

```toml
[build]
  cmd = "mage generate && go build -o ./tmp/server ./cmd/web"
  include_ext = ["go", "templ", "sql", "html", "css", "js"]
  exclude_regex = ["_test.go", "_templ.go", ".sql.go"]
  delay = 1000
```

### VS Code Settings

```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "files.associations": {
    "*.templ": "html"
  },
  "emmet.includeLanguages": {
    "templ": "html"
  }
}
```

## Common Issues

### Hot reload not working

```bash
# Check Air process
ps aux | grep air

# Restart development server
mage dev

# Check .air.toml file extensions
```

### SQLC generation fails

```bash
# Check SQL syntax
sqlc vet

# Verify schema is valid
sqlite3 :memory: '.read internal/store/schema.sql'
```

### Template compilation errors

```bash
# Check template syntax
templ generate

# Look for missing imports or syntax errors
```

This development guide provides everything needed for effective local development with hot reload, code generation, and comprehensive tooling.
