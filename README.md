# boring-go-web

Small Go starter app built with Echo, Templ, HTMX, PostgreSQL, and SQLC.

This repo currently gives you:

- Session-based login and registration
- A protected `/users` CRUD screen backed by Postgres
- CSRF protection, security headers, rate limiting, and structured errors
- Mage tasks for generate/build/lint/test
- Atlas migrations plus a schema bootstrap path for local bring-up

This repo does not give you:

- Roles or per-user authorization
- Password reset, email, or account recovery
- Metrics or `pprof` endpoints
- A polished design system
- A finished product architecture

## Quick Start

1. Make sure PostgreSQL is running locally.
2. Create a database and user:

```bash
createuser -P gowebserver
createdb -O gowebserver gowebserver
```

3. Create a local `.env` file:

```bash
cat > .env <<'EOF'
DATABASE_URL=postgres://gowebserver:your-password@localhost:5432/gowebserver?sslmode=disable
AUTH_COOKIE_SECURE=false
APP_ENVIRONMENT=development
APP_DEBUG=true
APP_LOG_LEVEL=debug
EOF
```

4. Install tools and generate assets:

```bash
mage setup
mage generate
```

5. Run the app:

```bash
mage run
```

The app listens on `http://localhost:8080`.

## Common Commands

```bash
mage dev            # Air hot reload
mage generate       # sqlc + templ + CSS
mage fmt            # goimports + gofmt + mod tidy
mage vet            # go vet
mage lint           # golangci-lint
mage ci             # generate + fmt + vet + lint + build
mage migrate        # apply Atlas migrations (requires atlas + .env)
mage migrateStatus  # show Atlas migration state
```

`mage migrateDown` is informational only. Atlas does not auto-rollback this repo.

## Docs

- [docs/README.md](/Users/sawyer/github/boring-go-web/docs/README.md)
- [docs/development.md](/Users/sawyer/github/boring-go-web/docs/development.md)
- [docs/api.md](/Users/sawyer/github/boring-go-web/docs/api.md)
- [docs/security.md](/Users/sawyer/github/boring-go-web/docs/security.md)
- [docs/architecture.md](/Users/sawyer/github/boring-go-web/docs/architecture.md)
- [docs/deployment.md](/Users/sawyer/github/boring-go-web/docs/deployment.md)

## Notes

- The checked-out directory is `boring-go-web`, but the current Go module path is still `github.com/dunamismax/go-web-server`.
- The canonical Atlas migration directory is top-level [`migrations/`](/Users/sawyer/github/boring-go-web/migrations). The duplicate `internal/store/migrations/` directory still exists and should be treated as leftover cleanup, not source of truth.
