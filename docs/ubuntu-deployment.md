# Ubuntu Deployment

This is the simplest supported deployment path in the repo today. It assumes Ubuntu, `systemd`, a locally reachable PostgreSQL instance, and the default `gowebserver` service names used by the scripts.

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

The build artifact is `bin/server`.

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

The deploy script expects `bin/server` and `.env` to exist in the repo root. It then:

- creates the `gowebserver` system user if needed
- copies `bin/server` to `/opt/gowebserver/bin/server`
- copies `.env` to `/opt/gowebserver/.env`
- installs [`scripts/gowebserver.service`](../scripts/gowebserver.service)
- reloads `systemd`, enables the service, and restarts it

## 6. Put a Reverse Proxy in Front

Use Caddy or Nginx for TLS and public exposure. The repo does not ship a managed proxy configuration that should be treated as production-ready.

## 7. Verify the Service

Useful commands after deployment:

```bash
sudo systemctl status gowebserver
sudo journalctl -u gowebserver -f
```

If you need different paths, service names, or a non-Ubuntu layout, edit [`scripts/deploy.sh`](../scripts/deploy.sh) and [`scripts/gowebserver.service`](../scripts/gowebserver.service) before installing.
