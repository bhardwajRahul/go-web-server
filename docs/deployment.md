# Deployment Guide

Production deployment instructions for the Go Web Server.

## Quick Deployment

### Build and Deploy

```bash
# Build production binary
mage build

# Deploy to server
scp bin/server user@server:/opt/app/
chmod +x /opt/app/server

# Run with environment variables
export APP_ENVIRONMENT=production
export DATABASE_URL=/opt/app/data/production.db
/opt/app/server
```

The binary is ~11MB with zero external dependencies.

## Configuration

### Environment Variables

**Production Settings:**

```bash
# Server
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Database
export DATABASE_URL=/opt/app/data/production.db
export DATABASE_RUN_MIGRATIONS=true

# Application
export APP_ENVIRONMENT=production
export APP_LOG_LEVEL=info
export APP_LOG_FORMAT=json
export APP_DEBUG=false

# Security
export SECURITY_ENABLE_CORS=true
export SECURITY_ALLOWED_ORIGINS=https://yourdomain.com
```

### Configuration File

`config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": "8080",
    "read_timeout": "10s",
    "write_timeout": "10s",
    "shutdown_timeout": "30s"
  },
  "database": {
    "url": "/opt/app/data/production.db",
    "run_migrations": true,
    "max_connections": 25
  },
  "app": {
    "environment": "production",
    "debug": false,
    "log_level": "info",
    "log_format": "json"
  },
  "security": {
    "enable_cors": true,
    "allowed_origins": ["https://yourdomain.com"]
  }
}
```

## Deployment Methods

### 1. Systemd Service

Create `/etc/systemd/system/go-web-server.service`:

```ini
[Unit]
Description=Go Web Server
After=network.target

[Service]
Type=exec
User=appuser
Group=appuser
WorkingDirectory=/opt/app
ExecStart=/opt/app/server
Restart=always
RestartSec=5

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/app/data

# Environment
Environment=APP_ENVIRONMENT=production
Environment=DATABASE_URL=/opt/app/data/production.db
Environment=APP_LOG_FORMAT=json

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-web-server
sudo systemctl start go-web-server
sudo systemctl status go-web-server
```

### 2. Docker Container

**Dockerfile:**

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/web

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

Build and run:

```bash
docker build -t go-web-server .
docker run -d \
  --name go-web-server \
  -p 8080:8080 \
  -v /opt/app/data:/data \
  -e DATABASE_URL=/data/production.db \
  -e APP_ENVIRONMENT=production \
  go-web-server
```

### 3. Docker Compose

**docker-compose.yml:**

```yaml
version: "3.8"

services:
  web:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - APP_ENVIRONMENT=production
      - DATABASE_URL=/data/production.db
      - APP_LOG_FORMAT=json
    restart: unless-stopped
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:8080/health",
        ]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - web
    restart: unless-stopped
```

## Reverse Proxy (Nginx)

**nginx.conf:**

```nginx
upstream go_web_server {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/cert.crt;
    ssl_certificate_key /etc/nginx/ssl/cert.key;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;

    location / {
        proxy_pass http://go_web_server;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 10s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    location /static/ {
        proxy_pass http://go_web_server;
        proxy_cache_valid 200 1d;
        add_header Cache-Control "public, immutable";
    }

    location /health {
        proxy_pass http://go_web_server;
        access_log off;
    }
}
```

## SSL/TLS Setup

### Let's Encrypt (Certbot)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d yourdomain.com

# Auto-renewal (add to crontab)
0 12 * * * /usr/bin/certbot renew --quiet
```

## Database Management

### Backup Script

**backup.sh:**

```bash
#!/bin/bash
DATABASE_FILE="/opt/app/data/production.db"
BACKUP_DIR="/opt/app/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Create backup with SQLite
sqlite3 $DATABASE_FILE ".backup $BACKUP_DIR/backup_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/backup_$DATE.db"

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.db.gz" -mtime +30 -delete

echo "Backup completed: backup_$DATE.db.gz"
```

Schedule with cron:

```bash
# Daily backup at 2 AM
0 2 * * * /opt/app/backup.sh
```

### Migration Management

```bash
# Check migration status
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db status

# Run migrations manually
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db up

# Rollback (if needed)
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db down
```

## Monitoring & Logging

### Health Checks

**health-check.sh:**

```bash
#!/bin/bash
HEALTH_URL="http://localhost:8080/health"
RESPONSE=$(curl -s -w "%{http_code}" "$HEALTH_URL")
HTTP_CODE="${RESPONSE: -3}"

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "Service is healthy"
    exit 0
else
    echo "Service is unhealthy: HTTP $HTTP_CODE"
    exit 1
fi
```

### Log Management

**With systemd:**

```bash
# View logs
sudo journalctl -u go-web-server -f

# Configure log rotation
sudo mkdir -p /etc/systemd/journald.conf.d
echo -e "[Journal]\nSystemMaxUse=1G\nMaxRetentionSec=30day" | \
sudo tee /etc/systemd/journald.conf.d/go-web-server.conf
```

**With Docker:**

```bash
# View logs
docker logs -f go-web-server

# Configure log driver in docker-compose.yml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Security Hardening

### Production Checklist

- [ ] Set `APP_ENVIRONMENT="production"`
- [ ] Disable debug mode (`APP_DEBUG="false"`)
- [ ] Configure specific CORS origins (not `*`)
- [ ] Use HTTPS with HSTS headers
- [ ] Set secure cookie settings
- [ ] Configure proper firewall rules
- [ ] Set database file permissions (`chmod 600`)
- [ ] Run as non-root user
- [ ] Enable comprehensive logging
- [ ] Set up security monitoring

### File Permissions

```bash
# Database permissions
chmod 600 /opt/app/data/production.db
chown appuser:appuser /opt/app/data/production.db

# Directory permissions
chmod 700 /opt/app/data/
chown appuser:appuser /opt/app/data/

# Binary permissions
chmod 755 /opt/app/server
chown appuser:appuser /opt/app/server
```

## Scaling Considerations

### Horizontal Scaling

**Load Balancer Setup:**

- Use Nginx, HAProxy, or cloud load balancer
- Configure health checks (`/health` endpoint)
- Session sticky not required (stateless design)

**Database Considerations:**

- SQLite suitable for moderate loads (thousands of concurrent users)
- Consider PostgreSQL for high-traffic scenarios
- Implement read replicas if needed

### Performance Tuning

**Database Connection Pool:**

```bash
export DATABASE_MAX_CONNECTIONS=25
export DATABASE_TIMEOUT=30s
```

**Application Profiling:**

```bash
# Enable pprof temporarily
export FEATURES_ENABLE_PPROF=true

# Profile endpoints
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof
```

## Troubleshooting

### Common Issues

**Service won't start:**

```bash
sudo systemctl status go-web-server
sudo journalctl -u go-web-server -n 50
ls -la /opt/app/
```

**Database issues:**

```bash
ls -la /opt/app/data/
sqlite3 /opt/app/data/production.db "PRAGMA integrity_check;"
```

**High memory usage:**

```bash
ps aux | grep server
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

This deployment guide ensures your Go Web Server runs reliably and securely in production with comprehensive monitoring and backup strategies.
