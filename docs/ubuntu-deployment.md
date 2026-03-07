# Ubuntu Deployment

This is the simplest supported path in the repo today.

## 1. Prepare PostgreSQL

```bash
sudo -u postgres createuser -P gowebserver
sudo -u postgres createdb -O gowebserver gowebserver
```

## 2. Build the App

```bash
mage setup
mage build
```

## 3. Create `.env`

```bash
cat > .env <<'EOF'
DATABASE_URL=postgres://gowebserver:your-password@localhost:5432/gowebserver?sslmode=disable
AUTH_COOKIE_SECURE=true
APP_ENVIRONMENT=production
APP_LOG_FORMAT=json
EOF
```

## 4. Optional: Apply Atlas Migrations

```bash
mage migrate
```

If you skip Atlas, the app will still bootstrap the current schema on startup.

## 5. Install the systemd Service

```bash
sudo ./scripts/deploy.sh
```

That script:

- creates the `gowebserver` system user if needed
- copies `bin/server` to `/opt/gowebserver/bin/server`
- copies `.env` to `/opt/gowebserver/.env`
- installs [`scripts/gowebserver.service`](/Users/sawyer/github/boring-go-web/scripts/gowebserver.service)

## 6. Put a Reverse Proxy in Front

Use Caddy or Nginx for TLS and public exposure. The repo does not ship a managed proxy configuration that should be treated as production-ready.
