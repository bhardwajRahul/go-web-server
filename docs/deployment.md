# Deployment

This repo includes a basic deployment path. It is a starting point, not a finished operations story.

## What Exists

- A build target: `mage build`
- A systemd unit file: [`scripts/gowebserver.service`](/Users/sawyer/github/boring-go-web/scripts/gowebserver.service)
- A helper script: [`scripts/deploy.sh`](/Users/sawyer/github/boring-go-web/scripts/deploy.sh)

## Minimum Deployment Shape

1. Provision PostgreSQL.
2. Build the binary with `mage build`.
3. Provide a `.env` file with at least `DATABASE_URL`.
4. Run Atlas migrations with `mage migrate` if you want explicit migration state.
5. Start the service behind a reverse proxy.

## Reality Check

- There is no built-in container workflow.
- There is no health-checked multi-instance setup.
- There is no metrics endpoint to scrape.
- The deploy script assumes Ubuntu + systemd and copies `bin/server` plus `.env`.

If you need something more serious than a single-host deployment, treat this repo as a code starting point and build the operational pieces intentionally.
