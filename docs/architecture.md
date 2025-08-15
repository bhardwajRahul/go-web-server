# Architecture Overview

System design and components of the Modern Go Stack web server.

## System Overview

```
Browser (HTMX + Tailwind CSS + DaisyUI)
         ↓
   Caddy (Reverse Proxy)
         ↓
  Go Web Server (Echo)
         ↓
 PostgreSQL (pgx/v5 driver)
```

## Request Flow

```
Request → Middleware Stack → Router → Handler → Store → Database
                ↓              ↓        ↓
         Security/Logging    CSRF    Templ
```

### Middleware Stack (15 layers)

1. **Recovery** - Panic recovery
2. **Security Headers** - Custom security headers
3. **Input Sanitization** - XSS/SQL injection prevention
4. **CSRF Protection** - Token validation
5. **Validation Errors** - Error conversion
6. **Timeout Handling** - Timeout error conversion
7. **Request ID** - Unique request tracing
8. **Prometheus Metrics** - Request metrics (optional)
9. **Structured Logging** - Request/response logging
10. **Echo Security** - XSS, HSTS headers
11. **CORS** - Cross-origin handling
12. **Rate Limiting** - 20 requests/minute per IP
13. **Timeout** - Request timeout enforcement
14. **Environment Context** - Environment to context
15. **Handler Execution** - Business logic

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Language** | Go 1.25+ | Latest performance & language features |
| **Server** | Echo v4 | High-performance HTTP framework with middleware |
| **Templates** | Templ v0.3.924 | Type-safe Go HTML components |
| **Frontend** | HTMX 2.x | Dynamic interactions without JavaScript complexity |
| **Styling** | Tailwind CSS + DaisyUI | Utility-first CSS with component library |
| **Authentication** | Session-based + Argon2id | Secure session auth with password hashing |
| **Database** | PostgreSQL + pgx/v5 | Enterprise-grade database with high-performance driver |
| **Queries** | SQLC v1.29.0 | Generate type-safe Go from SQL |
| **Validation** | go-playground/validator | Comprehensive input validation |
| **Logging** | slog | Structured logging with JSON output |
| **Config** | Koanf | Multi-source configuration management |
| **Migrations** | Atlas | Declarative schema management |
| **Build** | Mage | Go-based build automation |
| **Dev** | Air | Hot reload development server |
| **Assets** | Go Embed | Single binary with embedded resources |

## Project Structure

```
cmd/web/                     # Application entry point
├── main.go                  # Server setup, middleware stack, graceful shutdown
internal/
├── config/                  # Configuration management
│   └── config.go           # Koanf multi-source config with defaults
├── handler/                 # HTTP request handlers
│   ├── routes.go           # Route registration and static file serving
│   ├── auth.go            # Session authentication (login, register, logout)
│   ├── home.go            # Home page and health check endpoints
│   └── user.go            # User CRUD operations with HTMX
├── middleware/              # Security and validation middleware
│   ├── auth.go            # Session middleware and Argon2id password hashing
│   ├── csrf.go            # CSRF protection with token rotation
│   ├── errors.go          # Structured error handling and recovery
│   ├── metrics.go         # Prometheus metrics collection
│   ├── sanitize.go        # Input sanitization (XSS/SQL injection)
│   └── validation.go      # Request validation with go-playground/validator
├── store/                   # Database layer (SQLC generated)
│   ├── db.go              # Database connection and health checks
│   ├── models.go          # Generated Go structs from SQL schema
│   ├── queries.sql        # SQL queries for SQLC generation
│   ├── queries.sql.go     # Generated type-safe Go code from SQL
│   ├── schema.sql         # Database schema definition
│   ├── store.go           # Store interface and initialization
│   └── migrations/        # Atlas database migrations
│       ├── 20241231000001_initial_schema.sql
│       └── 20250815000001_add_sessions_and_passwords.sql
├── ui/                      # Embedded static assets
│   ├── embed.go           # Go embed directives for static files
│   └── static/            # Static web assets
│       ├── css/           # Stylesheets (Tailwind CSS, custom animations)
│       ├── js/            # JavaScript (HTMX)
│       └── favicon.ico    # Application favicon
└── view/                    # Templ templates
    ├── auth.templ         # Authentication forms (login/register)
    ├── home.templ         # Home page and demo components
    ├── users.templ        # User management interface
    └── layout/            # Layout components
        └── base.templ     # Base HTML layout with HTMX configuration
scripts/                     # Deployment and automation
├── deploy.sh              # Ubuntu deployment script
└── gowebserver.service    # SystemD service configuration
magefile.go                  # Mage build automation with comprehensive tasks
sqlc.yaml                    # SQLC configuration for code generation
```

## Data Flow

### HTMX Request Cycle

```
1. User Action → HTMX intercepts
2. AJAX request → Server middleware stack
3. Handler processes → Database query
4. Template renders → HTML fragment
5. HTMX swaps content → Triggers events
```

### Database Interaction

```
1. Handler validates input
2. Calls SQLC-generated store method
3. Store executes parameterized query
4. PostgreSQL processes and returns results
5. Results mapped to Go structs
6. Passed to template for rendering
```

## Security Architecture

### Defense in Depth

```
Input Layer:    Sanitization + Validation + Rate Limiting
App Layer:      CSRF + Security Headers + Error Handling
Data Layer:     Parameterized Queries + Input Sanitization
Transport:      HTTPS + HSTS + Secure Cookies
```

**Key Security Features:**

- Custom CSRF protection with token rotation
- Input sanitization for XSS/SQL injection
- Security headers (CSP, HSTS, X-Frame-Options)
- Rate limiting by IP address
- Structured error handling without information disclosure

## Build Architecture

### Single Binary Deployment

```
bin/server                # ~15MB executable
├── Embedded assets       # CSS, JS, templates
└── Minimal dependencies  # Requires PostgreSQL server
```

### Code Generation Pipeline

```
SQL changes → SQLC generates Go code
Template changes → Templ generates Go functions
CSS changes → Tailwind builds optimized stylesheet
Build → Single binary with embedded assets
```

## Performance Characteristics

**Memory Efficiency:**

- Compiled templates (no runtime parsing)
- Connection pooling for PostgreSQL
- Embedded assets (no file I/O)
- Minimal garbage collection

**Request Latency:**

- High-performance PostgreSQL driver
- Efficient middleware stack
- Type-safe database queries
- Minimal allocations

**Concurrency Model:**

- Goroutine per request
- Context-aware operations
- Database connection pooling
- Graceful shutdown

## Monitoring Architecture

**Request Tracing:**

- Unique request ID for correlation across logs
- Structured logging with slog (JSON format in production)
- Context propagation through middleware stack
- Error tracking with stack traces and request context

**Prometheus Metrics:**

- **HTTP Metrics**: Request duration, status codes, in-flight requests
- **Database Metrics**: Connection pool stats, query duration, operation counts
- **Application Metrics**: Version info, startup time, active users
- **HTMX Metrics**: HTMX-specific request tracking
- **Security Metrics**: CSRF token generation and validation failures
- **Business Metrics**: User creation, authentication events

**Health Monitoring:**

- Database connectivity checks
- Connection pool health
- System resource monitoring
- Configurable health check endpoints

**Error Handling:**

- Structured error responses with correlation IDs
- Production-safe error messages (no internal details exposed)
- Comprehensive error categorization (validation, authentication, etc.)
- Panic recovery with detailed logging

## Performance Characteristics

**Optimizations:**

- **Compiled Templates**: No runtime template parsing overhead
- **Connection Pooling**: Configurable PostgreSQL connection management
- **Embedded Assets**: All static files embedded in binary (no file I/O)
- **Type-Safe Queries**: SQLC eliminates reflection and runtime query parsing
- **Minimal Allocations**: Efficient request handling with object reuse
- **Context Cancellation**: Proper request timeout and cancellation handling

**Scalability Features:**

- **Horizontal Scaling**: Stateless design allows multiple instances
- **Database Scaling**: Connection pooling with configurable limits
- **Resource Management**: Memory and connection limits via SystemD
- **Graceful Shutdown**: Clean connection closure and request completion

This architecture provides production-ready performance, security, and operational visibility while maintaining Go's simplicity and developer experience.
