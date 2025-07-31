# Architecture Overview

System design and components of the Modern Go Web Server.

## Architecture Principles

1. **Pragmatic Simplicity** - Proven technologies over trendy ones
2. **Production-First** - Every decision optimizes for production deployment
3. **Developer Experience** - Minimize cognitive load, maximize productivity

## System Overview

```
Browser (HTMX + Pico.css)
         ↓
    Load Balancer
         ↓
   Go Web Server
         ↓
    SQLite Database
```

## Layer Architecture

### 1. Request Flow

```
Request → Middleware Stack → Router → Handler → Store → Database
```

**Middleware Order** (from `cmd/web/main.go:96-184`):

1. Recovery & Security Headers
2. Input Sanitization
3. CSRF Protection
4. Request ID & Logging
5. Rate Limiting & CORS
6. Timeout & Context

### 2. Component Structure

```
cmd/web/main.go           # Entry point & server setup
internal/
├── config/              # Koanf configuration management
├── handler/             # HTTP request handlers
├── middleware/          # Security & validation middleware
├── store/               # SQLC database layer
├── ui/                  # Embedded static assets
└── view/                # Templ templates
```

## Technology Stack

| Layer         | Technology                  | Purpose                         |
| ------------- | --------------------------- | ------------------------------- |
| **Server**    | Echo v4                     | HTTP framework with middleware  |
| **Templates** | Templ                       | Type-safe Go HTML templates     |
| **Frontend**  | HTMX 2.x                    | Dynamic interactions without JS |
| **Styling**   | Pico.css v2                 | Semantic CSS with themes        |
| **Database**  | SQLite + modernc.org/sqlite | Zero-CGO, embedded database     |
| **Queries**   | SQLC                        | Generate type-safe Go from SQL  |
| **Config**    | Koanf                       | Multi-source configuration      |
| **Build**     | Mage                        | Go-based build automation       |
| **Dev**       | Air                         | Hot reload development          |

## Data Flow

### 1. HTMX Request Cycle

```
1. User Action → HTMX intercepts
2. AJAX request → Server middleware stack
3. Handler processes → Database query
4. Template renders → HTML fragment
5. HTMX swaps content → Triggers events
```

### 2. Database Interaction

```
1. Handler validates input
2. Calls SQLC-generated store method
3. Store executes parameterized query
4. SQLite processes and returns results
5. Results mapped to Go structs
6. Passed to template for rendering
```

## Security Architecture

### Defense in Depth

```
Input Layer:    Sanitization + Validation + Rate Limiting
App Layer:      CSRF + Secure Headers + Error Handling
Data Layer:     Parameterized Queries + Input Sanitization
Transport:      HTTPS + HSTS + Secure Cookies
```

### Key Security Features

- **CSRF Protection**: `internal/middleware/csrf.go` - Custom implementation with token rotation
- **Input Sanitization**: `internal/middleware/sanitize.go` - XSS and SQL injection prevention
- **Security Headers**: CSP, HSTS, X-Frame-Options, X-XSS-Protection
- **Rate Limiting**: 20 requests/minute per IP with Echo middleware
- **Error Handling**: Structured errors without information disclosure

## Database Design

### Schema Management

```
internal/store/
├── migrations/          # Goose database migrations
├── schema.sql          # Current database schema
├── queries.sql         # Source SQL queries
├── queries.sql.go      # SQLC generated Go code
└── models.go           # SQLC generated models
```

### Type Safety Flow

```
1. Write SQL in queries.sql
2. Run `sqlc generate`
3. Get type-safe Go methods
4. Use in handlers without SQL injection risk
```

## Configuration Architecture

### Multi-Source Loading (Priority Order)

```go
1. Environment variables (highest)
2. Configuration files (JSON/YAML/TOML)
3. Default values in code (lowest)
```

### Production Overrides (`internal/config/config.go:108-114`)

```go
if environment == "production" {
    cfg.App.Debug = false
    cfg.App.LogFormat = "json"
    cfg.Security.AllowedOrigins = []string{}
    cfg.Database.RunMigrations = false
}
```

## Build Architecture

### Mage Build System

```
mage ci → generate + fmt + vet + lint + build
mage dev → generate + air (hot reload)
mage quality → vet + lint + vulncheck
```

### Asset Embedding

```go
//go:embed static/*
var StaticFiles embed.FS

// Single binary contains:
// - Go executable (~11MB)
// - Compiled templates
// - CSS/JS assets
// - Database schema
```

## Development Architecture

### Hot Reload System (Air)

```
File Change → Air detects → Generate code → Rebuild → Restart server
```

### Code Generation Pipeline

```
SQL changes → SQLC generates Go code
Template changes → Templ generates Go functions
Build → Single binary with embedded assets
```

## Production Architecture

### Single Binary Deployment

```
bin/server                # ~11MB executable
├── Embedded assets       # CSS, JS, templates
├── Database schema       # SQLite schema
└── Zero dependencies     # No external requirements
```

### Horizontal Scaling

```
Load Balancer
├── App Instance 1:8080
├── App Instance 2:8080
└── App Instance 3:8080
```

**Scaling Considerations:**

- SQLite suitable for moderate loads
- Consider PostgreSQL for high-traffic
- Stateless design enables easy horizontal scaling

## Error Handling Architecture

### Structured Error Types (`internal/middleware/errors.go`)

```go
type AppError struct {
    Type      ErrorType  // Categorized error types
    Code      int        // HTTP status code
    Message   string     // User-friendly message
    Details   any        // Additional context
    Internal  error      // Internal error (not exposed)
    RequestID string     # Request tracing
}
```

### Error Flow

```
Error occurs → AppError created → Context added →
Logged internally → Sanitized response → JSON sent to client
```

## Monitoring Architecture

### Request Tracing

```go
// Every request gets unique ID
requestID := c.Response().Header().Get(echo.HeaderXRequestID)

// Used in logs and error responses
slog.Info("request", "request_id", requestID, ...)
```

### Structured Logging

```go
// Development: Text format for readability
// Production: JSON format for log aggregation
```

## Performance Characteristics

### Memory Efficiency

- Compiled templates (no runtime parsing)
- Connection pooling ready
- Minimal garbage collection overhead
- Embedded assets (no file I/O)

### Request Latency

- Zero-CGO database driver
- Efficient middleware stack
- Type-safe database queries
- Minimal allocations in hot paths

This architecture provides production-ready performance and security while maintaining the simplicity that makes Go excellent for web development.
