# Security

This repo has a decent default baseline for a starter app. It does not have a full application security model.

## What Exists

### Session Authentication

- Sessions are managed with SCS and stored in PostgreSQL.
- Newly registered users get Argon2id password hashes.
- Protected routes use session middleware, not JWTs.

### CSRF Protection

- All state-changing routes go through custom CSRF middleware in [`internal/middleware/csrf.go`](/Users/sawyer/github/boring-go-web/internal/middleware/csrf.go).
- Tokens are checked against the `_csrf` cookie.
- Tokens rotate after successful state-changing requests.

### Output and Query Safety

- HTML rendering is done with Templ, which escapes values by default.
- Database access goes through SQLC-generated parameterized queries.
- Those two controls are the real primary defenses against XSS and SQL injection here.

### Request Normalization

- [`internal/middleware/sanitize.go`](/Users/sawyer/github/boring-go-web/internal/middleware/sanitize.go) now trims form/query values and strips NUL bytes.
- It does not try to “sanitize SQL” or pre-escape HTML before storage.
- That is deliberate. Pre-escaping stored data and mutating SQL-looking input is a good way to corrupt data while pretending to be security.

### Other Middleware

- Security headers via Echo secure middleware
- IP-based rate limiting with in-memory storage
- Structured error handling with request IDs

## What Does Not Exist

- No role-based authorization
- No per-record ownership checks
- No password reset or email verification
- No audit log
- No metrics-backed security monitoring
- No distributed/session revocation story beyond deleting sessions from the DB

## Current Risks

- Demo users without password hashes can still log in with arbitrary passwords.
- The app distinguishes only between “logged in” and “not logged in”.
- The rate limiter is in-memory, so it is per-process only.

If this repo becomes a real app, the next honest steps are removing passwordless demo logins, adding authorization rules, and deciding whether you want Atlas bootstrap-only, migrations-only, or both.
