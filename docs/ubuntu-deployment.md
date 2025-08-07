# Ubuntu SystemD Deployment Guide

Complete guide for deploying the Go web server on Ubuntu with systemd service management.

## Prerequisites

- Ubuntu 20.04+ server
- PostgreSQL 12+ installed locally
- Go 1.24+ (for building)
- Caddy or Nginx for reverse proxy
- Cloudflare DNS (optional)

## Database Setup

### Install PostgreSQL

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### Create Database and User

```bash
# Switch to postgres user
sudo -u postgres psql

# Create database and user
CREATE DATABASE gowebserver;
CREATE USER gowebserver WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE gowebserver TO gowebserver;

# Grant schema privileges
\c gowebserver
GRANT ALL ON SCHEMA public TO gowebserver;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gowebserver;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gowebserver;

\q
```

## Application Deployment

### 1. Build Application

```bash
# Clone repository
git clone https://github.com/your-username/go-web-server.git
cd go-web-server

# Install Go dependencies and tools
mage setup

# Create environment file
cp .env.example .env

# Edit environment file with your settings
nano .env
```

### 2. Environment Configuration

Update `.env` file:

```env
# Database
DATABASE_USER=gowebserver
DATABASE_PASSWORD=your_secure_password
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=gowebserver
DATABASE_URL=postgres://gowebserver:your_secure_password@localhost:5432/gowebserver?sslmode=disable

# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Features
FEATURES_ENABLE_METRICS=true

# Security
CSRF_SECRET_KEY=your-32-character-secret-key-here
```

### 3. Build and Deploy

```bash
# Build production binary
mage build

# Run database migrations
mage migrate

# Deploy using the deployment script
sudo ./scripts/deploy.sh
```

## Manual Deployment Steps

If you prefer manual deployment:

### 1. Create System User

```bash
# Create gowebserver system user
sudo useradd --system --shell /bin/false --home /opt/gowebserver --create-home gowebserver
```

### 2. Setup Directory Structure

```bash
# Create application directories
sudo mkdir -p /opt/gowebserver/{bin,logs}
sudo mkdir -p /var/log/gowebserver

# Set ownership
sudo chown -R gowebserver:gowebserver /opt/gowebserver
sudo chown -R gowebserver:gowebserver /var/log/gowebserver
```

### 3. Install Application

```bash
# Copy binary
sudo cp ./bin/server /opt/gowebserver/bin/
sudo chown gowebserver:gowebserver /opt/gowebserver/bin/server
sudo chmod 755 /opt/gowebserver/bin/server

# Copy environment file
sudo cp .env /opt/gowebserver/
sudo chown gowebserver:gowebserver /opt/gowebserver/.env
sudo chmod 600 /opt/gowebserver/.env
```

### 4. Install SystemD Service

```bash
# Copy service file
sudo cp ./scripts/gowebserver.service /etc/systemd/system/

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable gowebserver
sudo systemctl start gowebserver
```

## Reverse Proxy Setup

### Caddy Configuration

Create `/etc/caddy/Caddyfile`:

```caddy
your-domain.com {
    reverse_proxy localhost:8080
    
    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
        Referrer-Policy "strict-origin-when-cross-origin"
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'"
    }
    
    # Gzip compression
    encode gzip
    
    # Logging
    log {
        output file /var/log/caddy/access.log {
            roll_size 100mb
            roll_keep 10
        }
        format json
    }
}
```

### Nginx Configuration (Alternative)

Create `/etc/nginx/sites-available/gowebserver`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL configuration (use certbot for Let's Encrypt)
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # Reverse proxy
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }
}
```

## Monitoring and Maintenance

### Service Management

```bash
# Check service status
sudo systemctl status gowebserver

# Start/stop/restart service
sudo systemctl start gowebserver
sudo systemctl stop gowebserver
sudo systemctl restart gowebserver

# View logs
sudo journalctl -u gowebserver -f
sudo journalctl -u gowebserver --since today
sudo journalctl -u gowebserver --since "1 hour ago"
```

### Log Management

```bash
# View application logs
sudo tail -f /var/log/gowebserver/*.log

# Rotate logs (setup logrotate)
sudo nano /etc/logrotate.d/gowebserver
```

Add to logrotate:

```
/var/log/gowebserver/*.log {
    daily
    missingok
    rotate 30
    compress
    notifempty
    create 644 gowebserver gowebserver
    postrotate
        sudo systemctl reload gowebserver
    endscript
}
```

### Health Checks

```bash
# Application health endpoint
curl http://localhost:8080/health

# Prometheus metrics (if enabled)
curl http://localhost:8080/metrics
```

### Updates and Deployment

```bash
# Pull latest changes
git pull origin main

# Build new version
mage build

# Run any new migrations
mage migrate

# Deploy updated binary
sudo ./scripts/deploy.sh

# Or manually update
sudo systemctl stop gowebserver
sudo cp ./bin/server /opt/gowebserver/bin/
sudo systemctl start gowebserver
```

## Security Considerations

1. **Firewall**: Only allow necessary ports (80, 443, 22)
2. **Database**: Ensure PostgreSQL only accepts local connections
3. **Service User**: Application runs as non-privileged `gowebserver` user
4. **File Permissions**: Proper ownership and restricted permissions
5. **SSL/TLS**: Always use HTTPS in production
6. **Regular Updates**: Keep system and dependencies updated

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check service logs
sudo journalctl -u gowebserver -n 50

# Check binary permissions
ls -la /opt/gowebserver/bin/server

# Check environment file
sudo cat /opt/gowebserver/.env
```

**Database connection errors:**
```bash
# Test database connection
sudo -u postgres psql -d gowebserver -c "SELECT version();"

# Check PostgreSQL service
sudo systemctl status postgresql
```

**Permission errors:**
```bash
# Fix ownership
sudo chown -R gowebserver:gowebserver /opt/gowebserver
sudo chown -R gowebserver:gowebserver /var/log/gowebserver
```

## Performance Tuning

### PostgreSQL Optimization

Edit `/etc/postgresql/14/main/postgresql.conf`:

```conf
# Memory
shared_buffers = 256MB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connections
max_connections = 100

# Logging
log_statement = 'ddl'
log_min_duration_statement = 1000
```

### System Limits

Edit `/etc/security/limits.conf`:

```conf
gowebserver soft nofile 65536
gowebserver hard nofile 65536
```

This guide provides a complete Ubuntu deployment setup with systemd service management, perfect for traditional server deployments behind a reverse proxy with Cloudflare DNS.