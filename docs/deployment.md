# Deployment Guide

Production deployment for the Modern Go Stack using Caddy, Ubuntu, and Cloudflare.

## Quick Deployment

### Build and Deploy

```bash
# Build production binary
mage build

# Deploy to Ubuntu server
scp bin/server user@server:/opt/app/
chmod +x /opt/app/server

# Run with environment variables
export APP_ENVIRONMENT=production
export DATABASE_URL=/opt/app/data/production.db
export FEATURES_ENABLE_METRICS=true
/opt/app/server
```

The binary is ~11MB with zero external dependencies.

## Ubuntu Server Setup

### System Preparation

**Update system:**

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget unzip sqlite3
```

**Create application user:**

```bash
sudo useradd -m -s /bin/bash appuser
sudo mkdir -p /opt/app/{data,backups}
sudo chown -R appuser:appuser /opt/app
```

### Application Installation

**Deploy binary:**

```bash
# Copy binary
sudo cp bin/server /opt/app/
sudo chown appuser:appuser /opt/app/server
sudo chmod +x /opt/app/server

# Set secure permissions
sudo chmod 700 /opt/app/data
sudo chmod 755 /opt/app
```

## Configuration

### Environment Variables

Create `/opt/app/.env`:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DATABASE_URL=/opt/app/data/production.db
DATABASE_RUN_MIGRATIONS=true

# Application
APP_ENVIRONMENT=production
APP_LOG_LEVEL=info
APP_LOG_FORMAT=json
APP_DEBUG=false

# Security
SECURITY_ENABLE_CORS=true
SECURITY_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Features
FEATURES_ENABLE_METRICS=true
FEATURES_ENABLE_PPROF=false
```

### Configuration File

Optional `/opt/app/config.json`:

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
  },
  "features": {
    "enable_metrics": true,
    "enable_pprof": false
  }
}
```

## Systemd Service

### Service Configuration

Create `/etc/systemd/system/go-web-server.service`:

```ini
[Unit]
Description=Go Web Server - Modern Stack
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
ReadWritePaths=/opt/app/data
CapabilityBoundingSet=
AmbientCapabilities=
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

# Environment
EnvironmentFile=/opt/app/.env

[Install]
WantedBy=multi-user.target
```

### Service Management

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable go-web-server
sudo systemctl start go-web-server

# Check status
sudo systemctl status go-web-server

# View logs
sudo journalctl -u go-web-server -f
```

## Caddy Reverse Proxy

### Installation

**Install Caddy on Ubuntu:**

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

### Configuration

**Basic Caddyfile** (`/etc/caddy/Caddyfile`):

```caddyfile
yourdomain.com {
    reverse_proxy localhost:8080
    encode gzip

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
    }
}
```

**Advanced Caddyfile** with caching and monitoring:

```caddyfile
yourdomain.com {
    reverse_proxy localhost:8080

    # Enable compression
    encode gzip zstd

    # Cache static assets
    handle /static/* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=31536000"
    }

    # Metrics endpoint (protect in production)
    handle /metrics {
        reverse_proxy localhost:8080
        # Add IP whitelist or basic auth here
        @internal_ips remote_ip 10.0.0.0/8 192.168.0.0/16
        abort @internal_ips 403
    }

    # Health check
    handle /health {
        reverse_proxy localhost:8080
        header Cache-Control "no-cache, no-store, must-revalidate"
    }

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
        Permissions-Policy "geolocation=(), microphone=(), camera=()"
    }

    # Logging
    log {
        output file /var/log/caddy/yourdomain.com.log
        format json
    }
}
```

**Benefits of Caddy:**

- Automatic HTTPS with Let's Encrypt certificates
- HTTP/2 and HTTP/3 support out of the box
- Zero-configuration TLS
- Automatic certificate renewal
- Built-in health checks
- Easy configuration syntax

### Service Management

```bash
# Enable and start Caddy
sudo systemctl enable caddy
sudo systemctl start caddy

# Check configuration
sudo caddy validate --config /etc/caddy/Caddyfile

# Reload configuration
sudo systemctl reload caddy

# View logs
sudo journalctl -u caddy -f
```

## Cloudflare Integration

### DNS Setup

**Steps:**

1. Add your domain to Cloudflare
2. Update nameservers to Cloudflare's
3. Set up DNS records:

   ```
   A    @           your-server-ip
   A    www         your-server-ip
   ```

4. Enable proxy (orange cloud) for both records

### SSL/TLS Configuration

**Recommended Settings:**

- **SSL/TLS Mode**: "Full (strict)"
- **Always Use HTTPS**: Enabled
- **HSTS**: Enabled with max-age 31536000
- **Minimum TLS Version**: 1.2

### Performance Optimization

**Speed Settings:**

- **Auto Minify**: Enable CSS, JavaScript, HTML
- **Brotli Compression**: Enabled
- **Rocket Loader**: Enabled for JavaScript optimization
- **Browser Cache TTL**: 4 hours or higher

**Caching Rules:**

```
Rule: Static Assets
Match: *yourdomain.com/static/*
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  - Browser Cache TTL: 1 month
```

### Security Settings

**Firewall Rules:**

```
Expression: (http.request.uri.path contains "/metrics")
Action: Block
```

**Security Level**: Medium or High
**Challenge Passage**: 30 minutes
**Bot Fight Mode**: Enabled

### Page Rules (Optional)

**Static Assets Caching:**

```
URL: *yourdomain.com/static/*
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  - Browser Cache TTL: 1 month
```

**API Endpoints:**

```
URL: *yourdomain.com/api/*
Settings:
  - Cache Level: Bypass
  - Security Level: High
```

## Database Management

### Backup Strategy

**Automated Backup Script** (`/opt/app/backup.sh`):

```bash
#!/bin/bash
DATABASE_FILE="/opt/app/data/production.db"
BACKUP_DIR="/opt/app/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Create SQLite backup
sqlite3 $DATABASE_FILE ".backup $BACKUP_DIR/backup_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/backup_$DATE.db"

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.db.gz" -mtime +30 -delete

# Log backup completion
echo "$(date): Backup completed: backup_$DATE.db.gz" >> $BACKUP_DIR/backup.log
```

**Set permissions and schedule:**

```bash
sudo chown appuser:appuser /opt/app/backup.sh
sudo chmod +x /opt/app/backup.sh

# Add to appuser's crontab
sudo -u appuser crontab -e
# Add: 0 2 * * * /opt/app/backup.sh
```

### Database Maintenance

**Manual backup:**

```bash
sudo -u appuser sqlite3 /opt/app/data/production.db ".backup /opt/app/backups/manual_$(date +%Y%m%d).db"
```

**Database integrity check:**

```bash
sudo -u appuser sqlite3 /opt/app/data/production.db "PRAGMA integrity_check;"
```

**Migration management:**

```bash
# Check migration status
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db status

# Run migrations manually (if needed)
sudo -u appuser goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db up
```

## Monitoring & Logging

### Health Monitoring

**Health Check Script** (`/opt/app/health-check.sh`):

```bash
#!/bin/bash
HEALTH_URL="http://localhost:8080/health"
RESPONSE=$(curl -s -w "%{http_code}" "$HEALTH_URL")
HTTP_CODE="${RESPONSE: -3}"

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "$(date): Service is healthy"
    exit 0
else
    echo "$(date): Service is unhealthy: HTTP $HTTP_CODE"
    # Send alert (email, Slack, etc.)
    exit 1
fi
```

**Schedule health checks:**

```bash
# Add to cron for monitoring
# */5 * * * * /opt/app/health-check.sh >> /var/log/health-check.log 2>&1
```

### Log Management

**Systemd log configuration:**

```bash
sudo mkdir -p /etc/systemd/journald.conf.d
echo -e "[Journal]\nSystemMaxUse=1G\nMaxRetentionSec=30day" | \
sudo tee /etc/systemd/journald.conf.d/go-web-server.conf

sudo systemctl restart systemd-journald
```

**View application logs:**

```bash
# Real-time logs
sudo journalctl -u go-web-server -f

# Logs from last hour
sudo journalctl -u go-web-server --since "1 hour ago"

# JSON formatted logs
sudo journalctl -u go-web-server -o json
```

**Caddy logs:**

```bash
# Caddy access logs
sudo tail -f /var/log/caddy/yourdomain.com.log

# Caddy error logs
sudo journalctl -u caddy -f
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
- [ ] Run as non-root user (appuser)
- [ ] Enable comprehensive logging
- [ ] Set up security monitoring

### Firewall Configuration

**UFW (Ubuntu Firewall):**

```bash
# Enable firewall
sudo ufw enable

# Allow SSH (be careful!)
sudo ufw allow ssh

# Allow HTTP and HTTPS (Caddy)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Deny direct access to application port
sudo ufw deny 8080/tcp

# Check status
sudo ufw status verbose
```

### File Permissions

```bash
# Application binary
sudo chmod 755 /opt/app/server
sudo chown appuser:appuser /opt/app/server

# Database permissions
sudo chmod 600 /opt/app/data/production.db
sudo chown appuser:appuser /opt/app/data/production.db

# Directory permissions
sudo chmod 700 /opt/app/data/
sudo chown appuser:appuser /opt/app/data/

# Configuration files
sudo chmod 600 /opt/app/.env
sudo chown appuser:appuser /opt/app/.env
```

## Performance Tuning

### Application Configuration

**Database Connection Pool:**

```bash
export DATABASE_MAX_CONNECTIONS=25
export DATABASE_TIMEOUT=30s
```

**Server Timeouts:**

```json
{
  "server": {
    "read_timeout": "10s",
    "write_timeout": "10s",
    "shutdown_timeout": "30s"
  }
}
```

### System Optimization

**Kernel parameters** (`/etc/sysctl.conf`):

```
# Network optimizations
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 32768 134217728
net.ipv4.tcp_wmem = 4096 32768 134217728

# File descriptor limits
fs.file-max = 2097152
```

**Apply changes:**

```bash
sudo sysctl -p
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
sudo -u appuser sqlite3 /opt/app/data/production.db "PRAGMA integrity_check;"
```

**Caddy SSL issues:**

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo journalctl -u caddy -n 50
```

**High memory usage:**

```bash
ps aux | grep server
curl http://localhost:8080/debug/pprof/heap > heap.prof
```

This deployment guide ensures your Modern Go Stack application runs reliably and securely in production with Ubuntu, Caddy, and Cloudflare integration.
