# Development Guide

This guide covers setting up and working with the Modern Go Web Server in your local development environment.

## Prerequisites

### Required Software

- **Go 1.24+** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Mage** - Build automation tool

### Installing Mage

```bash
go install github.com/magefile/mage@latest
```

Verify installation:
```bash
mage -version
```

## Initial Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/go-web-server.git
cd go-web-server
```

### 2. Install Dependencies and Tools

```bash
# Download Go module dependencies
go mod tidy

# Install all development tools
mage setup
```

This installs:
- `templ` - Template compiler
- `sqlc` - SQL code generator  
- `air` - Hot reload development server
- `golangci-lint` - Comprehensive linter
- `govulncheck` - Vulnerability scanner
- `goimports` - Enhanced Go formatter
- `goose` - Database migration tool

### 3. Generate Code

```bash
# Generate SQLC queries and Templ templates
mage generate
```

### 4. Start Development Server

```bash
# Start with hot reload
mage dev

# Or build and run manually
mage run
```

The server will start at `http://localhost:8080`.

## Development Workflow

### Daily Development Commands

```bash
# Start development with hot reload
mage dev

# Generate code after SQL/template changes
mage generate

# Format all code
mage fmt

# Run quality checks
mage quality

# Clean build artifacts
mage clean
```

### Code Generation Workflow

**When you modify SQL:**
1. Edit `internal/store/queries.sql`
2. Run `mage generate` or let Air auto-rebuild
3. Generated code appears in `internal/store/queries.sql.go`

**When you modify templates:**
1. Edit `.templ` files in `internal/view/`
2. Run `mage generate` or let Air auto-rebuild  
3. Generated code appears as `*_templ.go` files

## Project Structure Deep Dive

```
go-web-server/
├── cmd/web/                    # Application entry point
│   └── main.go                 # Server setup and configuration
│
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   │   └── config.go           # Koanf-based config loading
│   │
│   ├── handler/                # HTTP request handlers
│   │   ├── home.go             # Home page handlers
│   │   ├── routes.go           # Route registration
│   │   └── user.go             # User CRUD handlers
│   │
│   ├── middleware/             # Custom middleware
│   │   ├── csrf.go             # CSRF protection
│   │   ├── errors.go           # Error handling
│   │   ├── sanitize.go         # Input sanitization
│   │   └── validation.go       # Request validation
│   │
│   ├── store/                  # Database layer
│   │   ├── db.go               # Database connection
│   │   ├── migrations/         # Database migrations
│   │   ├── models.go           # SQLC generated models
│   │   ├── queries.sql         # SQL queries (source)
│   │   ├── queries.sql.go      # SQLC generated code
│   │   ├── schema.sql          # Database schema
│   │   └── store.go            # Store implementation
│   │
│   ├── ui/                     # Static assets
│   │   ├── embed.go            # Go embed directives
│   │   └── static/             # Static files
│   │       ├── css/
│   │       └── js/
│   │
│   └── view/                   # Templates
│       ├── home.templ          # Home page templates
│       ├── home_templ.go       # Generated template code
│       ├── layout/             # Layout templates
│       ├── users.templ         # User management templates
│       └── users_templ.go      # Generated template code
│
├── docs/                       # Documentation
├── bin/                        # Compiled binaries
├── tmp/                        # Temporary files (dev)
├── .air.toml                   # Hot reload configuration
├── .golangci.yml               # Linter configuration
├── magefile.go                 # Build automation
├── sqlc.yaml                   # SQLC configuration
├── go.mod                      # Go module definition
└── go.sum                      # Go module checksums
```

## Database Development

### Working with Migrations

**Create new migration:**
```bash
# Create new migration file
goose -dir internal/store/migrations create add_user_avatar sql

# Edit the generated file
vim internal/store/migrations/YYYYMMDD_add_user_avatar.sql
```

**Migration file structure:**
```sql
-- +goose Up
ALTER TABLE users ADD COLUMN avatar_url TEXT;
CREATE INDEX idx_users_avatar ON users(avatar_url);

-- +goose Down  
DROP INDEX idx_users_avatar;
ALTER TABLE users DROP COLUMN avatar_url;
```

**Run migrations:**
```bash
# Apply all pending migrations
mage migrate

# Check migration status
mage migrateStatus

# Rollback last migration
mage migrateDown
```

### Writing SQL Queries

**1. Add queries to `internal/store/queries.sql`:**
```sql
-- name: GetActiveUsers :many
SELECT * FROM users 
WHERE is_active = 1 
ORDER BY created_at DESC;

-- name: GetUserWithStats :one
SELECT 
    u.*,
    COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.id = ?
GROUP BY u.id;
```

**2. Generate Go code:**
```bash
mage generate
# or
sqlc generate
```

**3. Use in handlers:**
```go
func (h *UserHandler) GetActiveUsers(c echo.Context) error {
    ctx := c.Request().Context()
    users, err := h.store.GetActiveUsers(ctx)
    if err != nil {
        return handleDatabaseError(err, c)
    }
    
    component := view.UserList(users)
    return component.Render(ctx, c.Response().Writer)
}
```

## Template Development

### Template Structure

**Base layout (`internal/view/layout/base.templ`):**
```go
package layout

templ Base(title string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{title} - Go Web Server</title>
        <link rel="stylesheet" href="/static/css/pico.min.css">
    </head>
    <body>
        <nav>
            <ul>
                <li><a href="/">Home</a></li>
                <li><a href="/users">Users</a></li>
            </ul>
        </nav>
        <main>
            { children... }
        </main>
        <script src="/static/js/htmx.min.js"></script>
    </body>
    </html>
}
```

**Page template:**
```go
package view

import "github.com/your-org/go-web-server/internal/view/layout"

templ Users() {
    @layout.Base("User Management") {
        <section>
            <h1>User Management</h1>
            @UserList()
        </section>
    }
}

templ UserList() {
    <div hx-get="/users/list" hx-trigger="load">
        Loading users...
    </div>
}
```

### HTMX Integration Patterns

**Form submission with HTMX:**
```go
templ UserForm(user *store.User) {
    <form hx-post="/users" 
          hx-target="#user-list" 
          hx-swap="innerHTML">
        <input type="hidden" name="csrf_token" value={ getCSRFToken() }/>
        
        <label>
            Name:
            <input type="text" name="name" value={ getUserName(user) } required/>
        </label>
        
        <label>
            Email:
            <input type="email" name="email" value={ getUserEmail(user) } required/>
        </label>
        
        <button type="submit">
            { getSubmitText(user) }
        </button>
    </form>
}
```

**Dynamic content updates:**
```go
templ UserRow(user store.User) {
    <tr id={ "user-" + strconv.FormatInt(user.ID, 10) }>
        <td>{ user.Name }</td>
        <td>{ user.Email }</td>
        <td>
            <button hx-delete={ "/users/" + strconv.FormatInt(user.ID, 10) }
                    hx-confirm="Delete this user?"
                    hx-target={ "#user-" + strconv.FormatInt(user.ID, 10) }
                    hx-swap="outerHTML">
                Delete
            </button>
        </td>
    </tr>
}
```

## Handler Development

### Handler Structure

```go
type UserHandler struct {
    store *store.Store
}

func NewUserHandler(s *store.Store) *UserHandler {
    return &UserHandler{store: s}
}

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
    
    // Return updated user list for HTMX
    users, err := h.store.ListUsers(ctx)
    if err != nil {
        return middleware.NewAppError(
            middleware.ErrorTypeInternal,
            http.StatusInternalServerError,
            "Failed to fetch users",
        ).WithContext(c).WithInternal(err)
    }
    
    component := view.UserList(users)
    return component.Render(ctx, c.Response().Writer)
}
```

### Error Handling Patterns

**Database errors:**
```go
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

**Validation errors:**
```go
if !isValidEmail(email) {
    return middleware.NewAppErrorWithDetails(
        middleware.ErrorTypeValidation,
        http.StatusBadRequest,
        "Invalid email format",
        map[string]string{"email": "Please provide a valid email address"},
    ).WithContext(c)
}
```

## Configuration Development

### Environment Variables

Create `.env` file for development:
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Database Configuration
DATABASE_URL=data/development.db
DATABASE_RUN_MIGRATIONS=true

# Application Configuration
APP_ENVIRONMENT=development
APP_DEBUG=true
APP_LOG_LEVEL=debug
APP_LOG_FORMAT=text

# Security Configuration (development)
SECURITY_ENABLE_CORS=true
SECURITY_ALLOWED_ORIGINS=*
```

### Configuration Files

**config.json (optional):**
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
  }
}
```

### Access Configuration in Code

```go
// In handlers
func (h *Handler) SomeMethod(c echo.Context) error {
    cfg := c.Get("config").(*config.Config)
    
    if cfg.App.Debug {
        slog.Debug("Debug information", "data", someData)
    }
    
    return nil
}

// Add config to context in main.go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        c.Set("config", cfg)
        return next(c)
    }
})
```

## Testing and Quality

### Running Quality Checks

```bash
# Format code
mage fmt

# Static analysis
mage vet

# Comprehensive linting
mage lint

# Security vulnerability scan
mage vulncheck

# Run all quality checks
mage quality

# Full CI pipeline
mage ci
```

### Custom Linting Rules

Edit `.golangci.yml` to customize linting:

```yaml
linters-settings:
  revive:
    rules:
      - name: exported
        severity: warning
        disabled: false
        arguments:
          - "checkPrivateReceivers"
          - "sayRepetitiveInsteadOfStutters"

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - revive
    - gosec
    - misspell
```

## Debugging

### Logging Configuration

**Development logging:**
```go
// Text format for easy reading
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

slog.SetDefault(logger)
```

**Structured debugging:**
```go
slog.Debug("Processing request",
    "method", c.Request().Method,
    "path", c.Request().URL.Path,
    "user_id", getUserID(c),
    "request_id", c.Response().Header().Get(echo.HeaderXRequestID))
```

### Database Debugging

**Query debugging:**
```sql
-- Add to queries.sql for debugging
-- name: DebugUser :one
SELECT *, 
       datetime(created_at) as created_at_formatted,
       datetime(updated_at) as updated_at_formatted
FROM users 
WHERE id = ?;
```

**SQLite debugging:**
```bash
# Connect to database directly
sqlite3 data/development.db

# Useful SQLite commands
.tables                          # List tables
.schema users                    # Show table schema  
.headers on                      # Show column headers
.mode column                     # Column display mode
SELECT * FROM users LIMIT 5;    # Query data
```

### Development Tools

**Air Configuration (`.air.toml`):**
```toml
[build]
  # Commands to run when building
  cmd = "mage generate && go build -o ./tmp/server ./cmd/web"
  
  # File extensions to watch
  include_ext = ["go", "templ", "sql", "html", "css", "js"]
  
  # Files to exclude from watching
  exclude_regex = ["_test.go", "_templ.go", ".sql.go"]
  
  # Delay before rebuilding
  delay = 1000
```

**VS Code Settings:**
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

## Common Development Tasks

### Adding a New Feature

1. **Plan the feature:**
   - Identify required database changes
   - Design the API endpoints
   - Plan the user interface

2. **Database changes:**
   ```bash
   # Create migration
   goose -dir internal/store/migrations create add_feature sql
   
   # Edit migration file
   # Run migration
   mage migrate
   ```

3. **Add SQL queries:**
   ```sql
   -- Add to internal/store/queries.sql
   -- name: CreateFeature :one
   INSERT INTO features (name, description) VALUES (?, ?) RETURNING *;
   ```

4. **Generate code:**
   ```bash
   mage generate
   ```

5. **Create handlers:**
   ```go
   // Add to internal/handler/feature.go
   func (h *FeatureHandler) CreateFeature(c echo.Context) error {
       // Implementation
   }
   ```

6. **Add routes:**
   ```go
   // In internal/handler/routes.go
   e.POST("/features", handlers.Feature.CreateFeature)
   ```

7. **Create templates:**
   ```go
   // Add to internal/view/feature.templ
   templ FeatureForm() {
       // Template implementation
   }
   ```

8. **Test the feature:**
   ```bash
   mage dev
   # Test in browser
   ```

### Debugging Common Issues

**Hot reload not working:**
```bash
# Check Air is running
ps aux | grep air

# Restart Air
mage dev

# Check file extensions in .air.toml
```

**SQLC generation fails:**
```bash
# Check SQL syntax in queries.sql
sqlc vet

# Verify schema.sql is valid
sqlite3 :memory: '.read internal/store/schema.sql'
```

**Template compilation errors:**
```bash
# Check template syntax
templ generate

# Look for missing imports or syntax errors
```

**CSRF token issues:**
```bash
# Check middleware order in main.go
# Verify forms include csrf_token field
# Check cookie settings in browser dev tools
```

This development guide provides everything you need to work effectively with the Modern Go Web Server. The combination of hot reload, code generation, and comprehensive tooling creates an excellent developer experience while maintaining production-ready code quality.