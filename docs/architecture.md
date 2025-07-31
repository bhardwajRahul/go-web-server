# Architecture Overview

This document provides a comprehensive overview of the Modern Go Web Server architecture, following The Modern Go Stack principles.

## Philosophy

The architecture is built on three core principles:

1. **Pragmatic Simplicity** - Choose proven technologies over trendy ones
2. **Production-First** - Every decision optimizes for production deployment
3. **Developer Experience** - Minimize cognitive load and maximize productivity

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client (Browser)                         │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐    │
│  │    HTML     │  │     HTMX     │  │    Pico.css     │    │
│  │ (Semantic)  │  │ (Interactivity)│  │   (Styling)     │    │
│  └─────────────┘  └──────────────┘  └─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                         HTTP/HTTPS
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Load Balancer / Reverse Proxy             │
│                     (Nginx/Cloudflare)                     │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Go Web Server                             │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                Middleware Stack                     │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │   Security  │ │ Sanitization │ │    CSRF     │  │    │
│  │  │   Headers   │ │     Input    │ │ Protection  │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │Rate Limiting│ │   Request    │ │   Error     │  │    │
│  │  │    & CORS   │ │   Logging    │ │  Handling   │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                Echo Router                          │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │    Static   │ │    Web       │ │     API     │  │    │
│  │  │   Assets    │ │   Routes     │ │   Routes    │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                Handler Layer                        │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │    Home     │ │     User     │ │    API      │  │    │
│  │  │   Handler   │ │   Handler    │ │  Handlers   │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                 View Layer                          │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │   Templ     │ │   Layout     │ │ Components  │  │    │
│  │  │ Templates   │ │  Templates   │ │  (Reusable) │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                Store Layer                          │    │
│  │                                                     │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────┐  │    │
│  │  │    SQLC     │ │  Database    │ │  Migration  │  │    │
│  │  │ Generated   │ │  Connection  │ │  Management │  │    │
│  │  │   Queries   │ │    Pool      │ │   (Goose)   │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      SQLite Database                       │
│                     (File-based)                           │
└─────────────────────────────────────────────────────────────┘
```

## Layer-by-Layer Breakdown

### 1. Client Layer

**HTML (Semantic)**

- Clean, semantic HTML5 markup
- Accessibility-first approach
- Progressive enhancement ready

**HTMX (Interactivity)**

- Declarative JavaScript replacement
- Server-driven UI updates
- Seamless AJAX-like functionality without writing JavaScript

**Pico.css (Styling)**

- Minimal, semantic CSS framework
- Automatic dark/light theme support
- Beautiful defaults for HTML elements

### 2. Middleware Stack

**Security Middleware:**

```go
// Order matters - security first
e.Use(middleware.RecoveryMiddleware())
e.Use(middleware.SecurityHeadersMiddleware())
e.Use(middleware.Sanitize())
e.Use(middleware.CSRF())
e.Use(middleware.ValidationErrorMiddleware())
e.Use(middleware.TimeoutErrorHandler())
```

**Core Middleware:**

```go
// Request tracking and logging
e.Use(echomiddleware.RequestID())
e.Use(echomiddleware.RequestLogger())

// Performance and security
e.Use(echomiddleware.Secure())
e.Use(echomiddleware.CORS())
e.Use(echomiddleware.RateLimiter())
e.Use(echomiddleware.Timeout())
```

### 3. Routing Layer

**Route Organization:**

```go
// Static assets
e.GET("/static/*", staticHandler)

// Web pages (full HTML responses)
e.GET("/", handlers.Home.Index)
e.GET("/users", handlers.User.Users)

// HTMX fragments (partial HTML responses)
e.GET("/users/list", handlers.User.UserList)
e.GET("/users/form", handlers.User.UserForm)

// API endpoints (JSON responses)
api := e.Group("/api")
api.GET("/users/count", handlers.User.UserCount)

// Form submissions (state-changing operations)
e.POST("/users", handlers.User.CreateUser)    // CSRF required
e.PUT("/users/:id", handlers.User.UpdateUser) // CSRF required
```

### 4. Handler Layer

**Handler Structure:**

```go
type Handlers struct {
    Home *HomeHandler
    User *UserHandler
    // Additional handlers...
}

type UserHandler struct {
    store *store.Store  // Dependency injection
}
```

**Handler Responsibilities:**

- Request validation and sanitization
- Business logic coordination
- Error handling and response formatting
- Context management and request tracing

### 5. View Layer

**Template Architecture:**

```
internal/view/
├── layout/
│   └── base.templ          # Base layout with HTML structure
├── home.templ              # Home page templates
├── users.templ             # User management templates
└── components/             # Reusable components
    ├── forms.templ
    ├── tables.templ
    └── modals.templ
```

**Template Composition:**

```go
// Base layout
templ Base(title string) {
    <!DOCTYPE html>
    <html>
    <head>
        <title>{title}</title>
        <link rel="stylesheet" href="/static/css/pico.min.css">
    </head>
    <body>
        { children... }
        <script src="/static/js/htmx.min.js"></script>
    </body>
    </html>
}

// Page template using layout
templ Users() {
    @Base("Users") {
        @UsersContent()
    }
}
```

### 6. Store Layer

**Store Architecture:**

```go
type Store struct {
    *Queries    // SQLC generated queries (embedded)
    db *sql.DB  // Database connection
}

// Generated by SQLC from SQL files
type Queries struct {
    db DBTX  // Database interface
}
```

**Query Generation Workflow:**

```
1. Write SQL in queries.sql
2. Define schema in schema.sql
3. Run `sqlc generate`
4. Type-safe Go methods are generated
5. Use in handlers without SQL injection risk
```

### 7. Database Layer

**SQLite Benefits:**

- Zero configuration
- ACID compliance
- Excellent performance for read-heavy workloads
- Perfect for single-server deployments
- Easy backup and replication

**Migration Management:**

```
internal/store/migrations/
├── 20241231000001_initial_schema.sql
├── 20241231000002_add_user_indexes.sql
└── 20241231000003_user_profile_updates.sql
```

## Data Flow

### 1. Request Processing Flow

```
1. Client Request
   ├── Static Asset? → Serve from embedded files
   ├── GET Request? → Render full page or fragment
   └── POST/PUT/DELETE? → Process with CSRF validation

2. Middleware Processing
   ├── Security headers added
   ├── Input sanitization applied
   ├── CSRF token validated (if required)
   ├── Rate limiting checked
   └── Request logging initiated

3. Route Matching
   ├── Echo router finds handler
   ├── Path parameters extracted
   └── Handler method invoked

4. Handler Processing
   ├── Request context propagated
   ├── Input validation performed
   ├── Business logic executed
   └── Response generated

5. Response Rendering
   ├── Template rendering (if HTML)
   ├── JSON serialization (if API)
   └── Error handling (if error)
```

### 2. Database Interaction Flow

```
1. Handler Method
   ├── Gets request context
   ├── Validates input parameters
   └── Calls store method

2. Store Method (SQLC Generated)
   ├── Prepares SQL statement
   ├── Binds parameters safely
   ├── Executes query
   └── Scans results into Go structs

3. Database Layer
   ├── SQLite processes query
   ├── Returns results
   └── Connection returned to pool

4. Response Processing
   ├── Results passed to template
   ├── Template rendered to HTML
   └── Response sent to client
```

### 3. HTMX Interaction Flow

```
1. User Interaction
   ├── Button click or form submission
   ├── HTMX intercepts event
   └── AJAX request sent to server

2. Server Processing
   ├── Same middleware stack applied
   ├── Handler processes request
   └── Returns HTML fragment

3. Client Update
   ├── HTMX receives HTML response
   ├── Swaps content in DOM
   └── Triggers custom events if specified

4. State Synchronization
   ├── Other page elements can listen for events
   ├── Counter updates, list refreshes, etc.
   └── Page stays in sync without full reload
```

## Configuration Architecture

### Multi-Source Configuration

```go
// Priority order (highest to lowest):
1. Environment variables
2. Configuration files (JSON/YAML/TOML)
3. Default values in code

// Loading order:
config := koanf.New(".")
config.Load(structs.Provider(defaults))      // 3. Defaults
config.Load(file.Provider("config.json"))    // 2. File
config.Load(env.Provider())                  // 1. Environment
```

### Environment-Specific Overrides

```go
// Production overrides automatically applied
if cfg.App.Environment == "production" {
    cfg.App.Debug = false
    cfg.App.LogFormat = "json"
    cfg.Security.AllowedOrigins = []string{}  // Remove wildcards
    cfg.Database.RunMigrations = false        // Manual migration control
}
```

## Build Architecture

### Mage Build System

```go
// Build dependency graph
Build -> Generate -> (generateSqlc, generateTempl)
CI -> Generate, Fmt, Vet, Lint, Build
Quality -> Vet, Lint, VulnCheck
```

### Asset Embedding

```go
//go:embed static/*
var StaticFiles embed.FS

// Single binary contains:
// - Go executable
// - HTML templates (compiled)
// - CSS stylesheets
// - JavaScript files
// - Database schema
// - Migration files
```

## Security Architecture

### Defense in Depth

```
1. Input Layer
   ├── Sanitization middleware
   ├── Validation middleware
   └── Rate limiting

2. Application Layer
   ├── CSRF protection
   ├── Secure headers
   └── Error handling that prevents information disclosure

3. Data Layer
   ├── Parameterized queries (SQLC)
   ├── Input sanitization
   └── Database file permissions

4. Transport Layer
   ├── HTTPS enforcement
   ├── HSTS headers
   └── Secure cookie settings
```

### Error Handling Architecture

```go
// Structured error types
type AppError struct {
    Type      ErrorType  // Categorized error types
    Code      int        // HTTP status code
    Message   string     // User-friendly message
    Details   any        // Additional context
    Internal  error      // Internal error (not exposed)
    RequestID string     // Request tracing
}

// Error response structure
type ErrorResponse struct {
    Type      ErrorType `json:"type"`
    Error     string    `json:"error"`
    Message   string    `json:"message"`
    Code      int       `json:"code"`
    RequestID string    `json:"request_id"`
    Timestamp string    `json:"timestamp"`
}
```

## Scalability Considerations

### Horizontal Scaling

**Load Balancer Setup:**

```nginx
upstream go_servers {
    server app1:8080;
    server app2:8080;
    server app3:8080;
}

server {
    location / {
        proxy_pass http://go_servers;
    }
}
```

**Database Considerations:**

- SQLite works well for moderate traffic
- For high scale, consider PostgreSQL with similar architecture
- Read replicas can be added with minimal code changes

### Vertical Scaling

**Resource Optimization:**

- Memory-efficient templates (compiled)
- Connection pooling for database
- Optimized static asset serving
- Minimal garbage collection overhead

## Development Architecture

### Hot Reload System

```toml
# .air.toml configuration
[build]
  cmd = "mage generate && go build -o ./tmp/server ./cmd/web"
  include_ext = ["go", "templ", "sql"]
  exclude_regex = ["_test.go", "_templ.go", ".sql.go"]
```

### Code Generation Pipeline

```
1. Schema Changes (schema.sql)
   ├── Database structure updated
   └── Migration files created

2. Query Changes (queries.sql)
   ├── SQLC generates type-safe Go code
   └── Store methods automatically updated

3. Template Changes (*.templ)
   ├── Templ generates Go functions
   └── Template compilation validated

4. Hot Reload Trigger
   ├── Air detects file changes
   ├── Rebuild process initiated
   └── Server restarted with new code
```

## Production Architecture

### Single Binary Deployment

```
Deployment Package:
├── server (binary ~11MB)
├── config.json (optional)
└── data/ (database directory)

Runtime Requirements:
├── No external dependencies
├── No framework installations
├── No package managers
└── Just the compiled binary
```

### Monitoring Integration

```go
// Request tracing
requestID := c.Response().Header().Get(echo.HeaderXRequestID)

// Structured logging
slog.Info("request processed",
    "method", c.Request().Method,
    "path", c.Request().URL.Path,
    "status", c.Response().Status,
    "duration", time.Since(start),
    "request_id", requestID)

// Error tracking
slog.Error("application error",
    "error", err,
    "request_id", requestID,
    "user_agent", c.Request().UserAgent(),
    "remote_ip", c.RealIP())
```

## Architecture Benefits

### Developer Experience

- **Fast feedback loop** with hot reload
- **Type safety** throughout the stack
- **Single language** (Go) for entire application
- **Minimal configuration** required

### Production Benefits

- **Single binary deployment** simplifies operations
- **Zero external dependencies** reduces failure points
- **Excellent performance** with minimal resource usage
- **Built-in security** features

### Maintenance Benefits

- **Code generation** reduces boilerplate
- **Structured logging** aids debugging
- **Comprehensive error handling** improves reliability
- **Migration system** handles schema evolution

This architecture provides a solid foundation for building production-ready web applications while maintaining the simplicity and performance characteristics that make Go an excellent choice for web development.
