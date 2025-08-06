# Architecture Overview

System design and components of the Modern Go Stack web server.

## Design Principles

1. **Pragmatic Simplicity** - Proven technologies over trendy ones
2. **Production-First** - Every decision optimizes for deployment and operations
3. **Developer Experience** - Minimize cognitive load, maximize productivity

## System Overview

```
Browser (HTMX 2.x + Pico.css v2)
         ↓
    Caddy (Reverse Proxy + HTTPS)
         ↓
   Cloudflare (CDN + Security)
         ↓
   Go Web Server (Echo v4)
         ↓
    SQLite Database (Pure Go)
         ↓
    Prometheus Metrics
```

## Request Flow Architecture

### 1. Request Processing Pipeline

```
Request → Middleware Stack → Router → Handler → Store → Database
                ↓                    ↓        ↓
           Prometheus          Templ      SQLC
            Metrics           Templates   Queries
```

**Middleware Order** (15 layers from `cmd/web/main.go`):

1. **Recovery** - Panic recovery with structured logging
2. **Security Headers** - Additional security headers
3. **Input Sanitization** - XSS and SQL injection prevention
4. **CSRF Protection** - Token validation and rotation
5. **Validation Errors** - Convert validation errors to structured responses
6. **Timeout Handling** - Timeout error conversion
7. **Request ID** - Unique request tracing
8. **Prometheus Metrics** - Request metrics collection (optional)
9. **Structured Logging** - Request/response logging
10. **Echo Security** - XSS, CSRF, HSTS headers
11. **CORS** - Cross-origin request handling (configurable)
12. **Rate Limiting** - 20 requests/minute per IP
13. **Timeout** - Request timeout enforcement
14. **Environment Context** - Add environment to context
15. **Handler Execution** - Business logic

### 2. Component Structure

```
cmd/web/main.go           # Entry point & server setup
internal/
├── config/              # Koanf multi-source configuration
├── handler/             # HTTP request handlers
├── middleware/          # Security & validation middleware
├── store/               # SQLC database layer
├── ui/                  # Embedded static assets
└── view/                # Templ templates
```

## Technology Stack Architecture

| Layer         | Technology                  | Purpose                         |
| ------------- | --------------------------- | ------------------------------- |
| **Server**    | Echo v4                     | HTTP framework with middleware  |
| **Templates** | Templ v0.3.924              | Type-safe Go HTML templates     |
| **Frontend**  | HTMX 2.0.6                  | Dynamic interactions without JS |
| **Styling**   | Pico.css v2                 | Semantic CSS with themes        |
| **Database**  | SQLite + modernc.org/sqlite | Zero-CGO, embedded database     |
| **Queries**   | SQLC v1.29.0                | Generate type-safe Go from SQL  |
| **Config**    | Koanf                       | Multi-source configuration      |
| **Build**     | Mage                        | Go-based build automation       |
| **Dev**       | Air                         | Hot reload development          |

## Data Flow Architecture

### 1. HTMX Request Cycle

```
1. User Action → HTMX intercepts
2. AJAX request → Server middleware stack
3. Handler processes → Database query
4. Template renders → HTML fragment
5. HTMX swaps content → Triggers events
```

### 2. Database Interaction Flow

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

- **CSRF Protection**: Custom implementation with token rotation (`internal/middleware/csrf.go`)
- **Input Sanitization**: XSS and SQL injection prevention (`internal/middleware/sanitize.go`)
- **Security Headers**: CSP, HSTS, X-Frame-Options, X-XSS-Protection
- **Rate Limiting**: 20 requests/minute per IP with memory store
- **Error Handling**: Structured errors without information disclosure (`internal/middleware/errors.go`)

## Database Architecture

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
1. Environment variables (highest priority)
2. Configuration files (JSON/YAML/TOML)
3. Default values in code (lowest priority)
```

### Production Overrides

Automatic production mode detection in `internal/config/config.go`:

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
Caddy Load Balancer
├── App Instance 1:8080
├── App Instance 2:8080
└── App Instance 3:8080
```

**Scaling Considerations:**

- SQLite suitable for moderate loads (thousands of concurrent users)
- Consider PostgreSQL for high-traffic scenarios
- Stateless design enables easy horizontal scaling
- Session data stored in cookies (not server memory)

## Error Handling Architecture

### Structured Error Types

From `internal/middleware/errors.go`:

```go
type AppError struct {
    Type      ErrorType  // Categorized error types
    Code      int        // HTTP status code
    Message   string     // User-friendly message
    Details   any        // Additional context
    Internal  error      // Internal error (not exposed)
    RequestID string     // Request tracing
    Timestamp string     // Error timestamp
    Path      string     // Request path
    Method    string     // HTTP method
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

### Metrics Collection

**Prometheus Integration:**

- HTTP request metrics (duration, status, method)
- Database connection and query metrics
- HTMX-specific interaction tracking
- CSRF token generation and validation metrics
- User activity and business metrics
- Application health and uptime tracking

## Performance Characteristics

### Memory Efficiency

- **Compiled templates** - No runtime parsing overhead
- **Connection pooling** - SQLite connection management
- **Minimal garbage collection** - Careful memory allocation
- **Embedded assets** - No file I/O for static content

### Request Latency

- **Zero-CGO database driver** - Pure Go SQLite implementation
- **Efficient middleware stack** - Optimized processing order
- **Type-safe database queries** - No reflection overhead
- **Minimal allocations** - Optimized hot paths

### Concurrency Model

- **Goroutine per request** - Standard Go HTTP server model
- **Context-aware operations** - Proper cancellation support
- **Database connection pooling** - Managed by Go's sql package
- **Graceful shutdown** - Clean connection termination

## Template Architecture

### Templ Integration

```go
// Type-safe component definition
templ UserList(users []User) {
    <table>
        for _, user := range users {
            @UserRow(user)
        }
    </table>
}

// Compiled to efficient Go code
func UserList(users []User) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Optimized rendering code
    })
}
```

### HTMX Integration Pattern

```go
// Server-side event triggering
c.Response().Header().Set("HX-Trigger", "userCreated")

// Client-side reactive updates
hx-trigger="userCreated from:body"
```

This architecture provides production-ready performance and security while maintaining the simplicity and developer experience that makes Go excellent for web development.
