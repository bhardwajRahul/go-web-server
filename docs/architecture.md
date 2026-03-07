# Architecture

## Shape

```text
Browser
  -> Echo routes + middleware
  -> handlers
  -> store (SQLC)
  -> PostgreSQL
```

The repo is a small monolith. There is one binary, one Postgres database, and one main demo domain model: users.

## Request Flow

1. Echo receives the request.
2. Middleware applies recovery, security headers, request normalization, CSRF, request IDs, logging, rate limiting, and timeout handling.
3. Session middleware loads the current user, if any.
4. Handlers validate input, call the store, and render Templ views or JSON.

## Route Split

- Public: home, demo, health, login, registration, static assets
- Protected: profile, user CRUD, user count API

## Repo Layout

- [`cmd/web/main.go`](/Users/sawyer/github/boring-go-web/cmd/web/main.go): app bootstrap and middleware stack
- [`internal/handler/`](/Users/sawyer/github/boring-go-web/internal/handler): route handlers
- [`internal/middleware/`](/Users/sawyer/github/boring-go-web/internal/middleware): auth, CSRF, error, validation, normalization
- [`internal/store/`](/Users/sawyer/github/boring-go-web/internal/store): SQLC queries, schema, pool setup
- [`internal/view/`](/Users/sawyer/github/boring-go-web/internal/view): Templ views
- [`internal/ui/static/`](/Users/sawyer/github/boring-go-web/internal/ui/static): embedded CSS, JS, favicon
- [`migrations/`](/Users/sawyer/github/boring-go-web/migrations): Atlas-managed SQL migrations

## Schema Sources

- [`internal/store/schema.sql`](/Users/sawyer/github/boring-go-web/internal/store/schema.sql) is the canonical schema definition used by Atlas.
- [`internal/store/store.go`](/Users/sawyer/github/boring-go-web/internal/store/store.go) contains a matching bootstrap schema for local startup.

The duplicate `internal/store/migrations/` directory still exists, but the app and docs should treat top-level [`migrations/`](/Users/sawyer/github/boring-go-web/migrations) as the source of truth.
