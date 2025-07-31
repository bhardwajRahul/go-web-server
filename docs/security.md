# Security Configuration

This document outlines the comprehensive security features and configuration options for the Modern Go Web Server.

## Security Architecture

The security model follows defense-in-depth principles with multiple layers of protection:

1. **Input Validation & Sanitization** - Prevents malicious input
2. **CSRF Protection** - Prevents cross-site request forgery
3. **Security Headers** - Protects against common web vulnerabilities
4. **Rate Limiting** - Prevents abuse and DoS attacks
5. **Structured Error Handling** - Prevents information disclosure
6. **Request Tracing** - Enables security monitoring

## CSRF Protection

### Overview

Cross-Site Request Forgery (CSRF) protection is automatically enabled for all state-changing operations.

### Configuration

**Default Configuration:**

```go
CSRFConfig{
    TokenLength:    32,                    // Token length in bytes
    TokenLookup:    "header:X-CSRF-Token,form:csrf_token",
    CookieName:     "_csrf",
    CookiePath:     "/",
    CookieSecure:   false,               // Set to true in HTTPS
    CookieHTTPOnly: true,
    CookieSameSite: http.SameSiteStrictMode,
    CookieMaxAge:   86400,               // 24 hours
}
```

**Production Configuration:**

```go
CSRFConfig{
    CookieSecure:   true,                // HTTPS only
    CookieSameSite: http.SameSiteStrictMode,
    CookieMaxAge:   3600,                // 1 hour for better security
}
```

### Token Management

**For HTML Forms:**

```html
<form method="POST" action="/users">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <!-- other form fields -->
</form>
```

**For HTMX Requests:**

```html
<button hx-post="/api/endpoint" hx-headers='{"X-CSRF-Token": "{{.CSRFToken}}"}'>
  Submit
</button>
```

**JavaScript/AJAX:**

```javascript
// Get token from cookie
function getCSRFToken() {
  return document.cookie
    .split("; ")
    .find((row) => row.startsWith("_csrf="))
    ?.split("=")[1];
}

// Include in request headers
fetch("/api/endpoint", {
  method: "POST",
  headers: {
    "X-CSRF-Token": getCSRFToken(),
    "Content-Type": "application/json",
  },
  body: JSON.stringify(data),
});
```

### Skip CSRF for Specific Routes

```go
// In main.go, before applying CSRF middleware
apiGroup := e.Group("/api/public")
// Public API routes don't need CSRF protection
apiGroup.GET("/status", handlers.Status)

// Apply CSRF to protected routes
e.Use(middleware.CSRF())
```

## Input Sanitization

### Overview

All user input is automatically sanitized to prevent XSS, HTML injection, and basic SQL injection attacks.

### Sanitization Features

**HTML Sanitization:**

- Escapes HTML special characters (`<`, `>`, `&`, `"`, `'`)
- Removes potentially dangerous HTML tags
- Preserves safe content while blocking malicious code

**XSS Protection:**

- Removes JavaScript: URLs and event handlers
- Blocks script tags and other executable content
- Sanitizes CSS and style attributes
- Removes dangerous protocols (data:, vbscript:, etc.)

**SQL Injection Prevention:**

- Escapes SQL special characters
- Removes SQL comment patterns (`--`, `/*`, `*/`)
- Blocks dangerous SQL keywords in user input
- Note: Primary protection is parameterized queries

### Configuration

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
    SanitizeSQL:  false,  // Disable for API endpoints using JSON
    CustomSanitizers: []func(string) string{
        // Custom sanitization function
        func(input string) string {
            // Remove specific patterns
            return strings.ReplaceAll(input, "badword", "***")
        },
    },
}

e.Use(middleware.SanitizeWithConfig(config))
```

**Preset Configurations:**

```go
// For HTML content
e.Use(middleware.SanitizeWithConfig(middleware.HTMLSanitizeConfig))

// For form inputs
e.Use(middleware.SanitizeWithConfig(middleware.FormSanitizeConfig))

// For SQL-heavy applications
e.Use(middleware.SanitizeWithConfig(middleware.SQLSanitizeConfig))
```

### Manual Sanitization

```go
// Sanitize individual strings
clean := middleware.SanitizeString(userInput, middleware.DefaultSanitizeConfig)

// In handlers
func (h *Handler) ProcessInput(c echo.Context) error {
    input := c.FormValue("user_input")

    // Additional custom sanitization
    input = strings.TrimSpace(input)
    input = middleware.SanitizeString(input, middleware.FormSanitizeConfig)

    // Process sanitized input
    return h.processCleanInput(input)
}
```

## Security Headers

### Applied Headers

The security headers middleware automatically applies:

```go
// Echo's built-in secure middleware
SecureConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeNosniff:    "nosniff",
    XFrameOptions:         "DENY",
    HSTSMaxAge:            31536000,  // 1 year
    ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'",
}

// Additional custom headers
"Referrer-Policy": "strict-origin-when-cross-origin"
"Permissions-Policy": "geolocation=(), microphone=(), camera=()"
"Cross-Origin-Opener-Policy": "same-origin"
"Cross-Origin-Embedder-Policy": "require-corp"
```

### Content Security Policy (CSP)

**Default CSP:**

```
default-src 'self';
style-src 'self' 'unsafe-inline';
script-src 'self' 'unsafe-inline'
```

**Stricter Production CSP:**

```go
// In production, use nonces for inline scripts
ContentSecurityPolicy: "default-src 'self'; style-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none';"
```

**Custom CSP Configuration:**

```go
secureConfig := echomiddleware.SecureConfig{
    ContentSecurityPolicy: "default-src 'self'; " +
                          "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
                          "font-src 'self' https://fonts.gstatic.com; " +
                          "img-src 'self' data: https:; " +
                          "script-src 'self'; " +
                          "connect-src 'self'; " +
                          "frame-ancestors 'none'; " +
                          "base-uri 'self'; " +
                          "object-src 'none';",
}

e.Use(echomiddleware.SecureWithConfig(secureConfig))
```

### HSTS Configuration

**Development:**

```go
HSTSMaxAge: 0,  // Disabled for HTTP development
```

**Production:**

```go
HSTSMaxAge: 63072000,  // 2 years
// Include subdomains and preload
HSTSExcludeSubdomains: false,
HSTSPreloadEnabled: true,
```

## CORS Configuration

### Default Settings

```go
CORSConfig{
    AllowOrigins: []string{"*"},              // Development only
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders: []string{"*"},
    MaxAge:       86400,                      // 24 hours
}
```

### Production Configuration

```go
// Environment-based CORS configuration
func getCORSConfig(env string) echomiddleware.CORSConfig {
    if env == "production" {
        return echomiddleware.CORSConfig{
            AllowOrigins: []string{
                "https://yourdomain.com",
                "https://www.yourdomain.com",
                "https://app.yourdomain.com",
            },
            AllowMethods: []string{
                http.MethodGet,
                http.MethodPost,
                http.MethodPut,
                http.MethodPatch,
                http.MethodDelete,
            },
            AllowHeaders: []string{
                "Origin",
                "Content-Type",
                "Accept",
                "Authorization",
                "X-CSRF-Token",
                "X-Requested-With",
            },
            MaxAge: 3600,  // 1 hour
        }
    }

    // Development configuration
    return echomiddleware.DefaultCORSConfig
}

// Apply in main.go
if cfg.Security.EnableCORS {
    e.Use(echomiddleware.CORSWithConfig(getCORSConfig(cfg.App.Environment)))
}
```

### Disable CORS

```bash
# Via environment variable
export SECURITY_ENABLE_CORS=false

# Via configuration file
{
  "security": {
    "enable_cors": false
  }
}
```

## Rate Limiting

### Default Configuration

```go
RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(20),  // 20 requests per minute
    IdentifierExtractor: func(c echo.Context) (string, error) {
        return c.RealIP(), nil  // Rate limit by IP address
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
    Store: echomiddleware.NewRateLimiterMemoryStore(100),  // 100 req/min per user
    IdentifierExtractor: func(c echo.Context) (string, error) {
        // Use user ID from session/JWT
        userID := getUserID(c)
        if userID != "" {
            return "user:" + userID, nil
        }
        return c.RealIP(), nil  // Fallback to IP
    },
})

e.Use(rateLimiter)
```

**Different Limits for Different Routes:**

```go
// Strict rate limiting for auth endpoints
authGroup := e.Group("/auth")
authGroup.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(5),  // 5 req/min
}))

// Normal rate limiting for API
apiGroup := e.Group("/api")
apiGroup.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(100),  // 100 req/min
}))
```

**Redis-based Rate Limiting:**

```go
import "github.com/go-redis/redis/v8"

redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

rateLimiter := echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: NewRedisRateLimiterStore(redisClient, 60),
})
```

## Error Handling Security

### Information Disclosure Prevention

The error handling system prevents sensitive information disclosure:

**Development vs Production:**

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

**Custom Error Responses:**

```go
// Safe error for external users
func (h *Handler) GetUser(c echo.Context) error {
    user, err := h.store.GetUser(ctx, id)
    if err != nil {
        // Log detailed error internally
        slog.Error("database error", "error", err, "user_id", id)

        // Return generic error to user
        return middleware.NewAppError(
            middleware.ErrorTypeNotFound,
            http.StatusNotFound,
            "User not found",  // No database details exposed
        ).WithContext(c)
    }

    return c.JSON(http.StatusOK, user)
}
```

### Error Logging Security

**Sensitive Data Filtering:**

```go
func sanitizeForLogging(data map[string]interface{}) map[string]interface{} {
    sensitive := []string{"password", "token", "secret", "key", "authorization"}

    cleaned := make(map[string]interface{})
    for k, v := range data {
        key := strings.ToLower(k)
        isSensitive := false

        for _, pattern := range sensitive {
            if strings.Contains(key, pattern) {
                isSensitive = true
                break
            }
        }

        if isSensitive {
            cleaned[k] = "[REDACTED]"
        } else {
            cleaned[k] = v
        }
    }

    return cleaned
}
```

## Database Security

### Query Security

**Parameterized Queries (SQLC):**

```sql
-- queries.sql - Always use parameters
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, name, bio) VALUES (?, ?, ?) RETURNING *;
```

**Generated type-safe code prevents SQL injection:**

```go
// This code is generated by SQLC and uses proper parameterization
user, err := h.store.GetUserByEmail(ctx, email)
```

### Database File Security

**File Permissions:**

```bash
# Set restrictive permissions
chmod 600 /opt/app/data/production.db
chown appuser:appuser /opt/app/data/production.db

# Directory permissions
chmod 700 /opt/app/data/
```

**Environment Configuration:**

```bash
# Use absolute paths
export DATABASE_URL="/opt/app/data/production.db"

# Disable foreign keys in development only
export DATABASE_FOREIGN_KEYS="true"
```

## Session Security

### Cookie Configuration

```go
// Secure cookie settings
cookie := &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   3600,                    // 1 hour
    HttpOnly: true,                    // Prevent XSS access
    Secure:   cfg.App.Environment == "production",  // HTTPS only in prod
    SameSite: http.SameSiteStrictMode, // CSRF protection
}
```

### Session Storage

**Memory-based (development):**

```go
sessions := make(map[string]*Session)
```

**Redis-based (production):**

```go
import "github.com/go-redis/redis/v8"

type RedisSessionStore struct {
    client *redis.Client
    ttl    time.Duration
}

func (s *RedisSessionStore) Set(sessionID string, data *Session) error {
    return s.client.Set(ctx, sessionID, data, s.ttl).Err()
}
```

## HTTPS Configuration

### TLS Settings

**Minimum TLS Version:**

```go
server.TLSConfig = &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    },
}
```

**Automatic HTTPS Redirect:**

```go
// Redirect HTTP to HTTPS
e.Pre(echomiddleware.HTTPSRedirect())

// Or use HSTS header
e.Use(echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
    HSTSMaxAge: 31536000,
}))
```

## Security Monitoring

### Request Tracing

Every request gets a unique ID for security monitoring:

```go
// Request ID middleware
e.Use(echomiddleware.RequestID())

// Access in handlers
func (h *Handler) SecureEndpoint(c echo.Context) error {
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)

    slog.Info("security event",
        "event", "sensitive_access",
        "request_id", requestID,
        "user_ip", c.RealIP(),
        "user_agent", c.Request().UserAgent())

    return nil
}
```

### Security Event Logging

```go
// Security event types
const (
    SecurityEventCSRFFailure     = "csrf_failure"
    SecurityEventRateLimitHit    = "rate_limit_exceeded"
    SecurityEventSuspiciousInput = "suspicious_input"
    SecurityEventAuthFailure     = "auth_failure"
)

func logSecurityEvent(eventType string, c echo.Context, details map[string]interface{}) {
    slog.Warn("security event",
        "event_type", eventType,
        "request_id", c.Response().Header().Get(echo.HeaderXRequestID),
        "remote_ip", c.RealIP(),
        "user_agent", c.Request().UserAgent(),
        "path", c.Request().URL.Path,
        "method", c.Request().Method,
        "details", details)
}
```

### Intrusion Detection

**Suspicious Pattern Detection:**

```go
func detectSuspiciousInput(input string) bool {
    suspiciousPatterns := []string{
        `<script`,
        `javascript:`,
        `union select`,
        `drop table`,
        `../../../`,
        `%3Cscript`,
    }

    lowerInput := strings.ToLower(input)
    for _, pattern := range suspiciousPatterns {
        if strings.Contains(lowerInput, pattern) {
            return true
        }
    }

    return false
}

// In middleware
func securityMonitoringMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Check all form values
            for key, values := range c.Request().Form {
                for _, value := range values {
                    if detectSuspiciousInput(value) {
                        logSecurityEvent(SecurityEventSuspiciousInput, c, map[string]interface{}{
                            "field": key,
                            "value_length": len(value),
                        })
                    }
                }
            }

            return next(c)
        }
    }
}
```

## Security Checklist

### Development

- [ ] Enable CSRF protection for all forms
- [ ] Use HTTPS in development when possible
- [ ] Test input sanitization with malicious payloads
- [ ] Verify error responses don't leak information
- [ ] Test rate limiting functionality

### Production

- [ ] Force HTTPS with HSTS headers
- [ ] Configure strict CORS origins
- [ ] Use secure cookie settings
- [ ] Enable comprehensive logging
- [ ] Set up security monitoring
- [ ] Configure Web Application Firewall (WAF)
- [ ] Regular security scans and updates
- [ ] Database backup encryption
- [ ] Implement intrusion detection
- [ ] Regular penetration testing

### Ongoing Maintenance

- [ ] Monitor security logs regularly
- [ ] Update dependencies for security patches
- [ ] Review and rotate secrets
- [ ] Conduct security audits
- [ ] Train team on security best practices
- [ ] Keep documentation updated
- [ ] Test incident response procedures

This security configuration provides comprehensive protection while maintaining the performance and simplicity principles of The Modern Go Stack.
