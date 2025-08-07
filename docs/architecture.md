# Architecture Overview

System design and components of the Modern Go Stack web server.

## System Overview

```
Browser (HTMX + Pico.css)
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
| **Server** | Echo v4 | HTTP framework with middleware |
| **Templates** | Templ v0.3.924 | Type-safe HTML templates |
| **Frontend** | HTMX 2.0.6 | Dynamic interactions |
| **Styling** | Pico.css v2 | Semantic CSS with themes |
| **Database** | PostgreSQL + pgx/v5 | High-performance PostgreSQL driver |
| **Queries** | SQLC v1.29.0 | Type-safe Go from SQL |
| **Build** | Mage | Go-based automation |
| **Dev** | Air | Hot reload development |

## Project Structure

```
cmd/web/main.go           # Entry point & server setup
internal/
├── config/              # Koanf configuration
├── handler/             # HTTP request handlers
├── middleware/          # Security & validation
├── store/               # SQLC database layer
├── ui/                  # Embedded static assets
└── view/                # Templ templates
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
bin/server                # ~14MB executable
├── Embedded assets       # CSS, JS, templates
└── Minimal dependencies  # Requires PostgreSQL server
```

### Code Generation Pipeline

```
SQL changes → SQLC generates Go code
Template changes → Templ generates Go functions
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
- Unique ID per request with structured logging

**Metrics Collection:**
- HTTP request metrics (duration, status, method)
- Database connection and query metrics
- Business metrics (user operations)

This architecture provides production-ready performance and security while maintaining simplicity and developer experience.
