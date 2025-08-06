# Security Guide

Comprehensive security features and implementation for the Modern Go Stack.

## Security Architecture

Defense-in-depth approach with multiple protection layers:

1. **Input Sanitization** - XSS and SQL injection prevention
2. **CSRF Protection** - Cross-site request forgery prevention
3. **Security Headers** - Browser security controls
4. **Rate Limiting** - DoS and abuse prevention
5. **Error Handling** - Information disclosure prevention
6. **Request Tracing** - Security monitoring and forensics

## CSRF Protection

### Implementation Overview

Custom CSRF middleware in `internal/middleware/csrf.go` protects all state-changing operations (POST/PUT/PATCH/DELETE).

### Configuration

**Default Settings:**

```go
CSRFConfig{
    TokenLength:    32,                           // 32-byte tokens
    TokenLookup:    "header:X-CSRF-Token,form:csrf_token",
    CookieName:     "_csrf",
    CookiePath:     "/",
    CookieSecure:   false,                       // Set true for HTTPS
    CookieHTTPOnly: true,
    CookieSameSite: http.SameSiteStrictMode,
    CookieMaxAge:   86400,                       // 24 hours
}
```

**Production Settings:**

```go
CSRFConfig{
    CookieSecure:   true,                        // HTTPS only
    CookieMaxAge:   3600,                        // 1 hour for better security
}
```

### Token Usage Patterns

**HTML Forms:**

```html
<form method="POST" action="/users">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <!-- form fields -->
</form>
```

**HTMX Requests:**

```html
<button hx-post="/users" hx-headers='{"X-CSRF-Token": "{{.CSRFToken}}"}'>
  Submit
</button>
```

**JavaScript/AJAX:**

```javascript
function getCSRFToken() {
  return document.cookie
    .split("; ")
    .find((row) => row.startsWith("_csrf="))
    ?.split("=")[1];
}

fetch("/api/endpoint", {
  method: "POST",
  headers: {
    "X-CSRF-Token": getCSRFToken(),
    "Content-Type": "application/json",
  },
  body: JSON.stringify(data),
});
```

### Security Features

- **Constant-time comparison** - Prevents timing attacks
- **Token rotation** - New token generated on each state-changing request
- **Multiple extraction sources** - Header, form, or query parameter
- **Secure cookie settings** - HttpOnly, SameSite, Secure flags
- **Cryptographically secure tokens** - Using `crypto/rand`

## Input Sanitization

### Automatic Protection

All user input sanitized via `internal/middleware/sanitize.go`:

**HTML Sanitization:**

```go
// Escapes dangerous HTML characters
input = html.EscapeString(input)
```

**XSS Protection:**

```go
dangerous := []string{
    "javascript:", "vbscript:", "data:", "blob:",
    "<script", "</script>", "<iframe", "onload=", "onclick="
}
```

**SQL Injection Prevention:**

```go
sqlComments := []string{"--", "/*", "*/", "#"}
dangerousPatterns := []string{
    "union select", "drop table", "delete from", "exec("
}
```

### Configuration Options

**Default Configuration:**

```go
SanitizeConfig{
    SanitizeHTML: true,
    SanitizeSQL:  true,
    SanitizeXSS:  true,
}
```

**Custom Sanitization:**

```go
config := middleware.SanitizeConfig{
    SanitizeHTML: true,
    SanitizeXSS:  true,
    SanitizeSQL:  false,  // Disable for JSON APIs
    CustomSanitizers: []func(string) string{
        func(input string) string {
            return strings.ReplaceAll(input, "badword", "***")
        },
    },
}
```

## Security Headers

### Applied Headers

Comprehensive security headers via Echo middleware and custom additions:

```go
// Echo's SecureMiddleware
XSSProtection:         "1; mode=block"
ContentTypeNosniff:    "nosniff"
XFrameOptions:         "DENY"
HSTSMaxAge:            31536000    // 1 year
ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'"

// Additional custom headers
"Referrer-Policy": "strict-origin-when-cross-origin"
"Permissions-Policy": "geolocation=(), microphone=(), camera=()"
"Cross-Origin-Opener-Policy": "same-origin"
"Cross-Origin-Embedder-Policy": "require-corp"
```

### Content Security Policy (CSP)

**Current CSP (HTMX Compatible):**

```
default-src 'self';
style-src 'self' 'unsafe-inline';    # Required for Pico.css
script-src 'self' 'unsafe-inline'    # Required for HTMX inline attributes
```

**Stricter Production CSP:**

```go
ContentSecurityPolicy: "default-src 'self'; " +
                      "style-src 'self'; " +
                      "script-src 'self'; " +
                      "object-src 'none'; " +
                      "base-uri 'self'; " +
                      "frame-ancestors 'none';"
```

Note: `'unsafe-inline'` needed for HTMX attributes like `hx-post`. Consider using nonces for stricter security.

### HSTS Configuration

**Development:**

```go
HSTSMaxAge: 0  // Disabled for HTTP development
```

**Production:**

```go
HSTSMaxAge: 63072000  // 2 years
```

## CORS Configuration

### Default Settings

```go
CORSConfig{
    AllowOrigins: []string{"*"},              # Development only!
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders: []string{"*"},
    MaxAge:       86400,                      # 24 hours
}
```

### Production Configuration

```go
if env == "production" {
    return echomiddleware.CORSConfig{
        AllowOrigins: []string{
            "https://yourdomain.com",
            "https://www.yourdomain.com",
        },
        AllowMethods: []string{
            http.MethodGet,
            http.MethodPost,
            http.MethodPut,
            http.MethodPatch,
            http.MethodDelete,
        },
        AllowHeaders: []string{
            "Origin", "Content-Type", "Accept",
            "Authorization", "X-CSRF-Token", "X-Requested-With",
        },
        MaxAge: 3600,  // 1 hour
    }
}
```

## Rate Limiting

### Default Configuration

```go
RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(20),  // 20 req/min
    IdentifierExtractor: func(c echo.Context) (string, error) {
        return c.RealIP(), nil  // Rate limit by IP
    },
    ErrorHandler: func(_ echo.Context, err error) error {
        return middleware.ErrTooManyRequests.WithInternal(err)
    },
}
```

### Custom Rate Limiting

**Per-User Rate Limiting:**

```go
rateLimiter := echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(100),
    IdentifierExtractor: func(c echo.Context) (string, error) {
        userID := getUserID(c)
        if userID != "" {
            return "user:" + userID, nil
        }
        return c.RealIP(), nil  // Fallback to IP
    },
})
```

**Route-Specific Limits:**

```go
// Strict limits for sensitive endpoints
authGroup := e.Group("/auth")
authGroup.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(5),  // 5 req/min
}))
```

## Error Handling Security

### Information Disclosure Prevention

Structured error handling in `internal/middleware/errors.go` prevents sensitive information leakage:

```go
// Development - detailed errors
if cfg.App.Environment == "development" {
    errorResp.Details = err.Details
    errorResp.Message = err.Message
}

// Production - sanitized errors
if cfg.App.Environment == "production" {
    if code >= 500 {
        errorResp.Details = nil
        errorResp.Message = "Internal server error"
    }
}
```

### Error Types

Categorized error types for consistent handling:

```go
type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeNotFound      ErrorType = "not_found"
    ErrorTypeInternal      ErrorType = "internal"
    ErrorTypeCSRF          ErrorType = "csrf"
    ErrorTypeRateLimit     ErrorType = "rate_limit"
    // ... more types
)
```

### Safe Error Responses

```go
// Safe error - no sensitive details exposed
return middleware.NewAppError(
    middleware.ErrorTypeNotFound,
    http.StatusNotFound,
    "User not found",  // Generic message
).WithContext(c)

// Internal error logged but not exposed
slog.Error("database connection failed", "error", err, "query", query)
```

## Database Security

### Query Security (SQLC)

All database queries use parameterized statements generated by SQLC:

```sql
-- queries.sql - Always parameterized
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, name, bio) VALUES (?, ?, ?) RETURNING *;
```

Generated Go code prevents SQL injection:

```go
// Type-safe, parameterized query
user, err := h.store.GetUserByEmail(ctx, email)
```

### Database File Security

```bash
# Restrictive file permissions
chmod 600 /opt/app/data/production.db
chown appuser:appuser /opt/app/data/production.db

# Directory permissions
chmod 700 /opt/app/data/
```

## Cookie Security

### Secure Cookie Settings

```go
cookie := &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   3600,                            // 1 hour
    HttpOnly: true,                            // Prevent XSS access
    Secure:   cfg.App.Environment == "production",  // HTTPS only in prod
    SameSite: http.SameSiteStrictMode,         // CSRF protection
}
```

### CSRF Cookie Configuration

```go
// CSRF cookies use strict settings
CookieHTTPOnly: true,                          // No JavaScript access
CookieSameSite: http.SameSiteStrictMode,       // Strict same-site policy
CookieSecure:   true,                          // HTTPS only in production
```

## Security Monitoring

### Request Tracing

Every request gets unique ID for security monitoring:

```go
// Request ID middleware
e.Use(echomiddleware.RequestID())

// Log security events with request ID
slog.Warn("security event",
    "event_type", "csrf_failure",
    "request_id", c.Response().Header().Get(echo.HeaderXRequestID),
    "remote_ip", c.RealIP(),
    "user_agent", c.Request().UserAgent(),
    "path", c.Request().URL.Path)
```

### Security Event Logging

```go
const (
    SecurityEventCSRFFailure     = "csrf_failure"
    SecurityEventRateLimitHit    = "rate_limit_exceeded"
    SecurityEventSuspiciousInput = "suspicious_input"
)

func logSecurityEvent(eventType string, c echo.Context, details map[string]interface{}) {
    slog.Warn("security event",
        "event_type", eventType,
        "request_id", c.Response().Header().Get(echo.HeaderXRequestID),
        "remote_ip", c.RealIP(),
        "details", details)
}
```

### Intrusion Detection

```go
func detectSuspiciousInput(input string) bool {
    suspiciousPatterns := []string{
        "<script", "javascript:", "union select",
        "drop table", "../../../", "%3Cscript",
    }

    lowerInput := strings.ToLower(input)
    for _, pattern := range suspiciousPatterns {
        if strings.Contains(lowerInput, pattern) {
            return true
        }
    }
    return false
}
```

## Security Checklist

### Development

- [ ] CSRF protection enabled for all forms
- [ ] Input sanitization tested with malicious payloads
- [ ] Error responses don't leak sensitive information
- [ ] Rate limiting functionality verified
- [ ] Security headers properly configured

### Production

- [ ] `APP_ENVIRONMENT="production"` set
- [ ] Debug mode disabled (`APP_DEBUG="false"`)
- [ ] CORS configured with specific origins (not `*`)
- [ ] HTTPS enforced with HSTS headers
- [ ] Secure cookie settings enabled
- [ ] Database file permissions restricted (`chmod 600`)
- [ ] Application runs as non-root user
- [ ] Comprehensive security logging enabled
- [ ] Security monitoring and alerting configured
- [ ] Regular security scans scheduled

### Ongoing Security

- [ ] Monitor security logs regularly
- [ ] Update dependencies for security patches
- [ ] Review and rotate secrets periodically
- [ ] Conduct security audits
- [ ] Test incident response procedures
- [ ] Keep security documentation current

This security implementation provides comprehensive protection while maintaining the performance and simplicity of the Modern Go Stack.
