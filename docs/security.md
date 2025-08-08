# Security Guide

Comprehensive security implementation for the Modern Go Stack.

## Security Architecture

Defense-in-depth approach with multiple protection layers:

1. **Input Sanitization** - XSS and SQL injection prevention
2. **CSRF Protection** - Cross-site request forgery prevention with token rotation
3. **JWT Authentication** - Secure token-based authentication with bcrypt password hashing
4. **Security Headers** - Browser security controls and content security policy
5. **Rate Limiting** - DoS and abuse prevention (20 requests/minute per IP)
6. **Input Validation** - Comprehensive request validation with go-playground/validator
7. **Error Handling** - Information disclosure prevention with structured errors
8. **Request Tracing** - Security monitoring with unique request IDs

## CSRF Protection

### Implementation

Custom CSRF middleware in `internal/middleware/csrf.go` protects all state-changing operations (POST, PUT, PATCH, DELETE).

**Configuration:**

```go
CSRFConfig{
    TokenLength:    32,                                    // 32-byte random tokens
    TokenLookup:    "header:X-CSRF-Token,form:csrf_token", // Multiple token sources
    CookieName:     "_csrf",                               // Cookie name
    CookieHTTPOnly: true,                                  // Prevent XSS access
    CookieSameSite: http.SameSiteStrictMode,              // CSRF protection
    CookieMaxAge:   86400,                                 // 24 hours
}
```

**Features:**
- **Token Rotation**: New token generated for each request
- **Multiple Sources**: Supports header and form-based tokens
- **Constant-Time Validation**: Prevents timing attacks
- **Secure Cookies**: HTTPOnly and SameSite attributes
- **Automatic HTMX Integration**: JavaScript automatically includes tokens

### Usage Examples

**Form-based CSRF (Traditional):**

```html
<form method="POST" action="/users">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <input type="text" name="name" required />
  <button type="submit">Create User</button>
</form>
```

**HTMX with Automatic CSRF:**

```html
<!-- CSRF token automatically included by JavaScript -->
<button hx-post="/users" hx-vals='{"name": "John Doe"}'>
  Create User
</button>
```

**Manual Header-based CSRF:**

```html
<button 
  hx-post="/users" 
  hx-headers='{"X-CSRF-Token": "{{.CSRFToken}}"}' 
  hx-vals='{"name": "John Doe"}'>
  Create User
</button>
```

## JWT Authentication

### Implementation

Secure JWT implementation in `internal/middleware/auth.go` with:

**Security Features:**
- HMAC-SHA256 signing (configurable signing keys)
- Secure cookie storage with HTTPOnly and SameSite attributes
- Configurable token expiration (24 hours default)
- Password hashing with bcrypt (cost 12)
- Token validation on each request
- Automatic cookie clearing on logout

**Configuration:**

```go
AuthConfig{
    SigningKey:      []byte("your-secret-key"), // From JWT_SECRET env var
    TokenDuration:   24 * time.Hour,            // Token expiration
    RefreshDuration: 7 * 24 * time.Hour,        // Refresh token duration
    Issuer:          "go-web-server",           // JWT issuer
    CookieName:      "auth_token",              // Cookie name
    CookieSecure:    true,                      // HTTPS only (false in dev)
    CookieHTTPOnly:  true,                      // Prevent XSS access
}
```

### Usage Examples

**Protecting Routes:**

```go
// Apply JWT middleware to protected routes
protected := e.Group("/api")
protected.Use(middleware.JWTMiddleware(authService))
protected.GET("/profile", handlers.Profile)
```

**Optional Authentication:**

```go
// Optional authentication (doesn't require login)
e.Use(middleware.OptionalJWTMiddleware(authService))
```

**Getting Current User:**

```go
func ProfileHandler(c echo.Context) error {
    user, exists := middleware.GetCurrentUser(c)
    if !exists {
        return c.Redirect(http.StatusFound, "/login")
    }
    // Use authenticated user...
}
```

## Input Sanitization

### XSS Prevention

Custom sanitization middleware in `internal/middleware/sanitize.go`:

**Features:**
- HTML entity escaping
- JavaScript protocol removal
- Event handler removal  
- Script tag filtering
- Dangerous pattern detection

**Configuration:**

```go
SanitizeConfig{
    SanitizeHTML: true,  // HTML entity escaping
    SanitizeXSS:  true,  // XSS pattern removal
    SanitizeSQL:  true,  // Basic SQL injection prevention
}
```

**Automatic Sanitization:**
- All form values automatically sanitized
- POST form values cleaned
- Custom sanitization functions supported

### SQL Injection Prevention

**Primary Defense - Parameterized Queries:**

All database operations use SQLC-generated code with parameterized queries:

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, name, bio) 
VALUES ($1, $2, $3)
RETURNING *;
```

**Secondary Defense - Input Sanitization:**

Additional SQL injection patterns filtered by sanitization middleware:
- SQL comment removal (`--`, `/*`, `#`)
- Dangerous keyword detection (`UNION SELECT`, `DROP TABLE`, etc.)
- Quote escaping as fallback

## Security Headers

### Implementation

Comprehensive security headers in `main.go` and `internal/middleware/errors.go`:

**Echo Security Middleware:**

```go
echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeNosniff:    "nosniff", 
    XFrameOptions:         "DENY",
    HSTSMaxAge:            31536000, // 1 year
    ContentSecurityPolicy: "default-src 'self'; ...",
})
```

**Custom Security Headers:**

```go
// Additional security headers
c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
```

### Content Security Policy

Restrictive CSP policy balancing security and functionality:

```
default-src 'self';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://fonts.gstatic.com;
script-src 'self' 'unsafe-inline' 'unsafe-eval';
img-src 'self' data:;
connect-src 'self';
font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com;
```

**Policy Rationale:**
- `'unsafe-inline'` for styles: Required for Pico.css and dynamic theme switching
- `'unsafe-eval'` for scripts: Required for HTMX functionality
- External fonts: Google Fonts for typography
- Data URIs: For inline images/icons

## Rate Limiting

### Implementation

IP-based rate limiting with configurable limits:

**Configuration:**

```go
echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
    Store: echomiddleware.NewRateLimiterMemoryStore(20), // 20 req/min
    IdentifierExtractor: func(c echo.Context) (string, error) {
        return c.RealIP(), nil // Rate limit by client IP
    },
    ErrorHandler: func(_ echo.Context, err error) error {
        return middleware.ErrTooManyRequests.WithInternal(err)
    },
})
```

**Features:**
- Memory-based storage (suitable for single instance)
- Real IP detection (works behind proxies)
- Configurable limits per endpoint
- Graceful error handling
- Prometheus metrics integration

## Input Validation

### Comprehensive Validation

Using `go-playground/validator` with custom rules:

**Built-in Validations:**
- Email format validation
- Password strength requirements
- String length limits
- Required field validation
- URL format validation

**Custom Password Validation:**

```go
// Password must be 8+ chars with uppercase, lowercase, and numbers
func passwordValidator(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    return len(password) >= 8 &&
        strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
        strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
        strings.ContainsAny(password, "0123456789")
}
```

**Usage Example:**

```go
type RegisterRequest struct {
    Email           string `json:"email" validate:"required,email"`
    Name            string `json:"name" validate:"required,min=2,max=100"`
    Password        string `json:"password" validate:"required,password"`
    ConfirmPassword string `json:"confirm_password" validate:"required"`
}
```

## Error Handling & Information Disclosure Prevention

### Structured Error Responses

Comprehensive error handling in `internal/middleware/errors.go`:

**Error Categories:**
- Validation errors (400)
- Authentication errors (401)  
- Authorization errors (403)
- Not found errors (404)
- Rate limit errors (429)
- Internal server errors (500)

**Production Safety:**
- Internal error details never exposed to clients
- Generic error messages in production
- Detailed logging for debugging
- Request correlation IDs for tracing

**Error Response Format:**

```json
{
  "type": "validation",
  "error": "Bad Request",
  "message": "Validation failed",
  "details": [
    {"field": "email", "message": "invalid email format"}
  ],
  "code": 400,
  "request_id": "req-123456",
  "timestamp": "1640995200"
}
```

## Monitoring & Alerting

### Security Metrics

Prometheus metrics for security monitoring:

```go
// CSRF protection metrics
csrfTokensGenerated    // Total tokens generated
csrfValidationFailures // Failed CSRF validations

// Authentication metrics  
userLoginAttempts      // Total login attempts
userLoginFailures      // Failed login attempts
userRegistrations      // New user registrations

// Request security metrics
requestsBlocked        // Requests blocked by rate limiting
validationFailures     // Input validation failures
```

### Security Event Logging

Structured logging for security events:

```go
slog.Warn("Failed login attempt",
    "email", email,
    "remote_ip", c.RealIP(),
    "user_agent", c.Request().UserAgent(),
    "request_id", requestID)
```

## Security Checklist

### Development
- [ ] All forms include CSRF tokens
- [ ] Passwords hashed with bcrypt
- [ ] Input validation on all endpoints
- [ ] SQL queries use parameters (SQLC)
- [ ] Error messages don't leak information
- [ ] Security headers configured
- [ ] Rate limiting active

### Production
- [ ] HTTPS enabled (TLS certificates)
- [ ] Secure JWT signing keys
- [ ] Database access restricted
- [ ] Security headers verified
- [ ] Rate limits appropriate for traffic
- [ ] Monitoring and alerting configured
- [ ] Log retention policies set
- [ ] Regular security updates applied

## Security Best Practices

1. **JWT Management:**
   - Use strong, random signing keys (256+ bits)
   - Rotate signing keys regularly
   - Set appropriate token expiration times
   - Store tokens in HTTPOnly cookies

2. **Password Security:**
   - Enforce strong password requirements
   - Use bcrypt with appropriate cost (12+)
   - Implement account lockout after failed attempts
   - Consider 2FA for sensitive applications

3. **Database Security:**
   - Use parameterized queries exclusively
   - Implement connection pooling limits
   - Regular security updates for PostgreSQL
   - Database access limited to application user

4. **Monitoring:**
   - Monitor failed authentication attempts
   - Alert on unusual traffic patterns
   - Log security events with correlation IDs
   - Regular security audit reviews

This security implementation provides enterprise-grade protection while maintaining usability and performance.