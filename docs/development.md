# Development

## Local Setup

1. Make sure PostgreSQL is running.
2. Create a database and user:

```bash
createuser -P gowebserver
createdb -O gowebserver gowebserver
```

3. Create `.env` in the repo root:

```bash
cat > .env <<'EOF'
DATABASE_URL=postgres://gowebserver:your-password@localhost:5432/gowebserver?sslmode=disable
AUTH_COOKIE_SECURE=false
APP_ENVIRONMENT=development
APP_DEBUG=true
APP_LOG_LEVEL=debug
EOF
```

4. Install tools and build generated assets:

```bash
mage setup
mage generate
```

5. Start the app:

```bash
mage dev
```

Use `mage run` if you want a plain build-and-run without Air.

## Daily Commands

```bash
mage dev
mage generate
mage fmt
mage vet
mage lint
mage ci
```

## Database Notes

- The app calls `store.InitSchema()` on startup so a fresh local database can boot without Atlas.
- The canonical schema file is [`internal/store/schema.sql`](/Users/sawyer/github/boring-go-web/internal/store/schema.sql).
- The canonical migration directory is [`migrations/`](/Users/sawyer/github/boring-go-web/migrations).
- Use Atlas when you want explicit migration state:

```bash
mage migrate
mage migrateStatus
```

`mage migrateDown` does not roll back changes. It prints guidance only.

## What To Regenerate

Run `mage generate` after changing:

- [`internal/store/queries.sql`](/Users/sawyer/github/boring-go-web/internal/store/queries.sql)
- [`internal/view/*.templ`](/Users/sawyer/github/boring-go-web/internal/view)
- [`input.css`](/Users/sawyer/github/boring-go-web/input.css)

## Verification

For code changes in handlers, middleware, store, or views, the minimum useful checks are:

```bash
mage vet
mage lint
go test ./...
```

If Atlas is part of the change, also run `mage migrateStatus`.
