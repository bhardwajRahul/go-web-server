# Deployment Guide

Production deployment for the Modern Go Stack using Ubuntu, PostgreSQL, Caddy, and Cloudflare.

## Quick Deployment

```bash
# Build and deploy
mage build
scp bin/server user@server:/opt/app/

# Set environment and run
export APP_ENVIRONMENT=production
export DATABASE_URL=postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@localhost:5432/gowebserver?sslmode=disable
/opt/app/server
```

Binary is ~14MB with minimal external dependencies (requires PostgreSQL server).

## Ubuntu Server Setup

### System Preparation

```bash
# Update and install PostgreSQL
sudo apt update && sudo apt upgrade -y
sudo apt install -y postgresql postgresql-contrib

# Create application user and database
sudo useradd -m -s /bin/bash appuser
sudo mkdir -p /opt/app/{backups,logs}
sudo chown -R appuser:appuser /opt/app

# Setup PostgreSQL database
sudo -u postgres createuser -P gowebserver  # Set password when prompted
sudo -u postgres createdb -O gowebserver gowebserver
```

### Application Installation

```bash
# Deploy binary
sudo cp bin/server /opt/app/
sudo chown appuser:appuser /opt/app/server
sudo chmod +x /opt/app/server

# Set permissions
sudo chmod 755 /opt/app
sudo chmod 700 /opt/app/logs
```

## Configuration

### Environment Variables

Create `/opt/app/.env`:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DATABASE_URL=postgres://gowebserver:your_secure_password@localhost:5432/gowebserver?sslmode=disable
DATABASE_RUN_MIGRATIONS=true

# Application
APP_ENVIRONMENT=production
APP_LOG_LEVEL=info
APP_LOG_FORMAT=json
APP_DEBUG=false

# Security
SECURITY_ENABLE_CORS=true
SECURITY_ALLOWED_ORIGINS=https://yourdomain.com

# Features
FEATURES_ENABLE_METRICS=true
```

## Systemd Service

### Service Configuration

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

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/app/logs

# Environment
EnvironmentFile=/opt/app/.env

[Install]
WantedBy=multi-user.target
```

### Service Management

```bash
# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable go-web-server
sudo systemctl start go-web-server

# Check status and logs
sudo systemctl status go-web-server
sudo journalctl -u go-web-server -f
```

## Caddy Reverse Proxy

### Installation

```bash
# Install Caddy on Ubuntu
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install caddy
```

### Configuration

Create `/etc/caddy/Caddyfile`:

```caddyfile
yourdomain.com {
    reverse_proxy localhost:8080
    encode gzip

    # Cache static assets
    handle /static/* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=31536000"
    }

    # Health check (no cache)
    handle /health {
        reverse_proxy localhost:8080
        header Cache-Control "no-cache"
    }

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        Referrer-Policy "strict-origin-when-cross-origin"
    }

    # Logging
    log {
        output file /var/log/caddy/yourdomain.com.log
        format json
    }
}
```

**Benefits:**

- Automatic HTTPS with Let's Encrypt
- HTTP/2 and HTTP/3 support
- Zero-configuration TLS
- Automatic certificate renewal

### Service Management

```bash
# Enable and start Caddy
sudo systemctl enable caddy
sudo systemctl start caddy

# Validate configuration
sudo caddy validate --config /etc/caddy/Caddyfile

# Reload after changes
sudo systemctl reload caddy
```

## Cloudflare Integration

### DNS Setup

1. Add domain to Cloudflare
2. Update nameservers
3. Set DNS records:

   ```
   A    @           your-server-ip
   A    www         your-server-ip
   ```

4. Enable proxy (orange cloud)

### SSL/TLS Settings

- **SSL/TLS Mode**: "Full (strict)"
- **Always Use HTTPS**: Enabled
- **HSTS**: Enabled (max-age 31536000)
- **Minimum TLS Version**: 1.2

### Performance

- **Auto Minify**: Enable CSS, JavaScript, HTML
- **Brotli Compression**: Enabled
- **Browser Cache TTL**: 4 hours

### Security

- **Security Level**: Medium or High
- **Bot Fight Mode**: Enabled
- **Firewall Rules**:

  ```
  Expression: (http.request.uri.path contains "/metrics")
  Action: Block
  ```

## Database Management

### Automated PostgreSQL Backup

Create `/opt/app/backup.sh`:

```bash
#!/bin/bash
export PGPASSWORD="your_secure_password"
DATABASE_URL="postgres://gowebserver:your_secure_password@localhost:5432/gowebserver"
BACKUP_DIR="/opt/app/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
pg_dump $DATABASE_URL | gzip > "$BACKUP_DIR/backup_$DATE.sql.gz"

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

Set permissions and schedule:

```bash
sudo chown appuser:appuser /opt/app/backup.sh
sudo chmod +x /opt/app/backup.sh

# Add to crontab: 0 2 * * * /opt/app/backup.sh
sudo -u appuser crontab -e
```

### Database Maintenance

```bash
# Manual backup
export PGPASSWORD="your_secure_password"
pg_dump "postgres://gowebserver:your_secure_password@localhost:5432/gowebserver" | gzip > /opt/app/backups/manual_$(date +%Y%m%d).sql.gz

# Database connection test
psql "postgres://gowebserver:your_secure_password@localhost:5432/gowebserver" -c "SELECT version();"

# Performance stats
psql "postgres://gowebserver:your_secure_password@localhost:5432/gowebserver" -c "SELECT schemaname,tablename,n_tup_ins,n_tup_upd,n_tup_del FROM pg_stat_user_tables;"
```

## Monitoring

### Health Monitoring

Create `/opt/app/health-check.sh`:

```bash
#!/bin/bash
RESPONSE=$(curl -s -w "%{http_code}" "http://localhost:8080/health")
HTTP_CODE="${RESPONSE: -3}"

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "$(date): Service healthy"
    exit 0
else
    echo "$(date): Service unhealthy: HTTP $HTTP_CODE"
    exit 1
fi
```

Schedule: `*/5 * * * * /opt/app/health-check.sh >> /var/log/health-check.log 2>&1`

### Log Management

```bash
# View application logs
sudo journalctl -u go-web-server -f

# View Caddy logs
sudo tail -f /var/log/caddy/yourdomain.com.log
```

## Security Hardening

### Firewall

```bash
# Enable UFW
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 8080/tcp  # Block direct access
```

### File Permissions

```bash
# Application binary
sudo chmod 755 /opt/app/server
sudo chown appuser:appuser /opt/app/server

# Logs directory
sudo chmod 700 /opt/app/logs
sudo chown appuser:appuser /opt/app/logs

# Configuration
sudo chmod 600 /opt/app/.env
sudo chown appuser:appuser /opt/app/.env
```

## Troubleshooting

**Service won't start:**

```bash
sudo systemctl status go-web-server
sudo journalctl -u go-web-server -n 50
```

**Database issues:**

```bash
sudo systemctl status postgresql
psql "postgres://gowebserver:password@localhost:5432/gowebserver" -c "SELECT 1;"
```

**Caddy SSL issues:**

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo journalctl -u caddy -n 50
```

This deployment guide ensures reliable and secure production operation with Ubuntu, PostgreSQL, Caddy, and Cloudflare integration.
