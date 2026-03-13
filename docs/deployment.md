# Deployment

This repo includes a basic single-host deployment path. It is a starting point, not a full production platform.

## What Exists Today

- A build target: `mage build`
- A `systemd` unit file: [`scripts/gowebserver.service`](../scripts/gowebserver.service)
- A helper install script: [`scripts/deploy.sh`](../scripts/deploy.sh)
- A sample YAML config: [`docs/config.example.yaml`](config.example.yaml)

## Minimum Deployment Shape

1. Provision PostgreSQL.
2. Build the binary with `mage build`.
3. Provide configuration through `.env`, `config.yaml`, environment variables, or a mix of those.
4. Run Atlas migrations with `mage migrate` if you want explicit migration state.
5. Start the service behind a reverse proxy.

## Single-Host Checklist

- Build output lands at `bin/server`.
- The bundled Ubuntu deploy path copies `bin/server` and `.env` into `/opt/gowebserver/`.
- The service name, system user, and default paths in the deployment script are all `gowebserver`.
- `AUTH_COOKIE_SECURE` should be `true` anywhere you terminate TLS properly.
- `security.trusted_proxies` should stay empty unless you are behind proxies you control and have configured intentionally.

For the repo's concrete Ubuntu path, see [ubuntu-deployment.md](ubuntu-deployment.md).

## Reality Check

- There is no built-in container workflow.
- There is no health-checked multi-instance setup.
- There is no metrics endpoint to scrape.
- There is no zero-downtime deployment story in the repo.
- The deploy script assumes Ubuntu + `systemd` and copies `bin/server` plus `.env`.

If you need something more serious than a single-host deployment, treat this repo as a code starting point and build the operational pieces intentionally.
