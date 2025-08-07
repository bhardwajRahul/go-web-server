# Development Guide

Local development setup and workflow for the Modern Go Stack.

## Quick Setup

```bash
# Clone and setup
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server

# Install tools and dependencies
mage setup

# Start PostgreSQL database
docker compose up postgres -d

# Start development server with hot reload
mage dev
```

**Prerequisites:** Go 1.24+, Docker (for PostgreSQL)

Server runs at `http://localhost:8080`

## Daily Commands

```bash
mage dev          # Start with hot reload
mage generate     # Generate SQLC + Templ code
mage quality      # Run vet + lint + vulncheck
mage build        # Production binary
```

## Database Development

**PostgreSQL with Docker:**

```bash
docker compose up postgres -d    # Start PostgreSQL
mage migrate                    # Run migrations up
mage migrateDown               # Rollback last migration
mage migrateStatus             # Show migration status
```

**Creating migrations:**

```bash
goose -dir internal/store/migrations create feature_name sql
```

**Writing queries in `internal/store/queries.sql`:**

```sql
-- name: GetActiveUsers :many
SELECT * FROM users WHERE is_active = true ORDER BY created_at DESC;
```

## Template Development

**Base layout (`internal/view/layout/base.templ`):**

```go
package layout

templ Base(title string) {
    <!DOCTYPE html>
    <html lang="en" data-theme="dark">
        <head>
            <title>{title} - Go Web Server</title>
            <link rel="stylesheet" href="/static/css/pico.min.css">
            <script src="/static/js/htmx.min.js"></script>
        </head>
        <body>
            { children... }
        </body>
    </html>
}
```

**Page template with HTMX:**

```go
templ UserForm(user *store.User, token string) {
    <form hx-post="/users" hx-target="#user-list">
        <input type="hidden" name="csrf_token" value={token}/>
        <input type="text" name="name" required/>
        <button type="submit">Create User</button>
    </form>
}
```

## Handler Development

**Basic handler pattern:**

```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Validate input
    name := c.FormValue("name")
    if name == "" {
        return middleware.NewAppError(
            middleware.ErrorTypeValidation,
            http.StatusBadRequest,
            "Name is required",
        ).WithContext(c)
    }
    
    // Database operation
    user, err := h.store.CreateUser(ctx, store.CreateUserParams{
        Name: name,
        Email: c.FormValue("email"),
    })
    if err != nil {
        return middleware.NewAppError(
            middleware.ErrorTypeInternal,
            http.StatusInternalServerError,
            "Failed to create user",
        ).WithContext(c).WithInternal(err)
    }
    
    // Return HTML for HTMX
    component := view.UserRow(user)
    return component.Render(ctx, c.Response().Writer)
}
```

## Configuration

**Development environment variables:**

```bash
APP_ENVIRONMENT=development
APP_DEBUG=true
SERVER_PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/gowebserver?sslmode=disable
FEATURES_ENABLE_METRICS=true
```

## Debugging

**Database:**

```bash
# Connect to PostgreSQL in Docker
docker exec -it gowebserver-postgres psql -U user -d gowebserver

# Or using local psql client
psql postgres://user:password@localhost:5432/gowebserver

# Common commands
\dt                          # List tables
\d users                     # Describe users table
SELECT * FROM users LIMIT 5; # Query users
```

**Monitoring:**

```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## Common Issues

- **Hot reload not working:** Check `mage dev` and ensure Air is running
- **SQLC errors:** Run `sqlc vet` to check SQL syntax
- **Template errors:** Run `templ generate` to check syntax
