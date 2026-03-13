# Security

This repo has a decent starter baseline, but it is still a template. The security story is real enough to build on, not complete enough to call finished.

## What Exists

### Session Authentication

- Sessions are managed with SCS and stored in PostgreSQL.
- Protected routes use session middleware, not JWTs.
- Newly registered users get Argon2id password hashes.
- Accounts without a valid password hash are rejected during login.
- Session cookies are `HttpOnly`, `SameSite=Strict`, and use the configured `auth.cookie_secure` setting.

### CSRF Protection

- All state-changing routes go through custom CSRF middleware in [`internal/middleware/csrf.go`](../internal/middleware/csrf.go).
- Tokens are checked against the `_csrf` cookie.
- The middleware accepts tokens from `X-CSRF-Token` or `csrf_token`.
- Tokens rotate after successful state-changing requests.

### Output and Query Safety

- HTML rendering is done with Templ, which escapes values by default.
- Database access goes through SQLC-generated parameterized queries.
- Those two controls are the real primary defenses against XSS and SQL injection here.

### Request Normalization

- [`internal/middleware/sanitize.go`](../internal/middleware/sanitize.go) trims form/query values and strips NUL bytes.
- It does not try to “sanitize SQL” or pre-escape HTML before storage.
- That is deliberate. Pre-escaping stored data and mutating SQL-looking input is a good way to corrupt data while pretending to be security.

### Other Middleware

- Security headers via Echo secure middleware
- IP-based rate limiting with in-memory storage
- Structured error handling with request IDs
- Trusted proxy support is configurable, but should stay empty unless the app is actually behind proxies you control

## What Does Not Exist

- No role-based authorization
- No per-record ownership checks
- No password reset or email verification
- No audit log
- No metrics-backed security monitoring
- No distributed rate limiting
- No active use of the JWT config fields that still exist in config for future cleanup

## Current Risks

- The app distinguishes only between “logged in” and “not logged in”.
- The rate limiter is in-memory, so it is per-process only.
- Default CORS settings are permissive unless you tighten them in configuration.
- The session store is database-backed, but there is no deeper authorization model once a user is authenticated.

If this repo becomes a real app, the next honest steps are adding authorization rules, tightening CORS and deployment settings, and deciding whether you want Atlas bootstrap-only, migrations-only, or both.
