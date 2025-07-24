# Deployment Guide

This guide covers deploying the Modern Go Web Server in production environments.

## Quick Deployment

### 1. Build Production Binary

```bash
# Clone the repository
git clone https://github.com/your-org/go-web-server.git
cd go-web-server

# Install build tools
mage setup

# Build optimized production binary
mage build
```

The binary will be created at `bin/server` (~11MB, single file).

### 2. Deploy Binary

```bash
# Copy binary to production server
scp bin/server user@production-server:/opt/app/

# Set executable permissions
chmod +x /opt/app/server

# Run the server
/opt/app/server
```

## Configuration

### Environment Variables

Set these environment variables for production:

```bash
# Server Configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT="8080"

# Database Configuration  
export DATABASE_URL="/opt/app/data/production.db"
export DATABASE_RUN_MIGRATIONS="true"

# Application Configuration
export APP_ENVIRONMENT="production"
export APP_LOG_LEVEL="info"
export APP_LOG_FORMAT="json"
export APP_DEBUG="false"

# Security Configuration
export SECURITY_ENABLE_CORS="true"
export SECURITY_ALLOWED_ORIGINS="https://yourdomain.com"
```

### Configuration Files

Alternatively, create a configuration file:

**config.json:**
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
    "max_connections": 25,
    "timeout": "30s"
  },
  "app": {
    "environment": "production",
    "debug": false,
    "log_level": "info",
    "log_format": "json"
  },
  "security": {
    "enable_cors": true,
    "allowed_origins": ["https://yourdomain.com"],
    "trusted_proxies": ["127.0.0.1"]
  },
  "features": {
    "enable_metrics": true,
    "enable_pprof": false
  }
}
```

## Production Checklist

### Security
- [ ] Set `APP_ENVIRONMENT="production"`
- [ ] Disable debug mode (`APP_DEBUG="false"`)
- [ ] Configure specific CORS origins (not `*`)
- [ ] Use HTTPS in production
- [ ] Set secure cookie flags
- [ ] Configure proper firewall rules
- [ ] Review and update CSP headers

### Performance
- [ ] Use JSON logging (`APP_LOG_FORMAT="json"`)
- [ ] Configure appropriate log levels
- [ ] Set database connection limits
- [ ] Configure request timeouts
- [ ] Enable gzip compression (reverse proxy)

### Monitoring
- [ ] Set up log aggregation
- [ ] Configure health checks
- [ ] Monitor database file size
- [ ] Set up alerts for errors
- [ ] Track response times

### Backup
- [ ] Schedule database backups
- [ ] Test backup restoration
- [ ] Store backups securely
- [ ] Document recovery procedures

## Deployment Methods

### 1. Systemd Service

Create `/etc/systemd/system/go-web-server.service`:

```ini
[Unit]
Description=Go Web Server
Documentation=https://github.com/your-org/go-web-server
After=network.target

[Service]
Type=exec
User=appuser
Group=appuser
WorkingDirectory=/opt/app
ExecStart=/opt/app/server
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/app/data

# Environment
Environment=APP_ENVIRONMENT=production
Environment=DATABASE_URL=/opt/app/data/production.db
Environment=APP_LOG_FORMAT=json

[Install]
WantedBy=multi-user.target
```

**Enable and start:**
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
RUN apk add --no-cache git
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/web

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/internal/ui/static ./internal/ui/static

EXPOSE 8080
CMD ["./server"]
```

**Build and run:**
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
version: '3.8'

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
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
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

## Reverse Proxy Setup

### Nginx Configuration

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
        
        # Timeouts
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

## SSL/TLS Configuration

### Let's Encrypt with Certbot

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d yourdomain.com

# Auto-renewal (add to crontab)
0 12 * * * /usr/bin/certbot renew --quiet
```

### Manual SSL Setup

1. **Obtain SSL certificates** (from your CA)
2. **Install certificates:**
   ```bash
   sudo cp yourdomain.com.crt /etc/nginx/ssl/
   sudo cp yourdomain.com.key /etc/nginx/ssl/
   sudo chmod 600 /etc/nginx/ssl/*
   ```
3. **Update nginx configuration** (see above)

## Database Management

### Backup Strategy

**Daily backup script:**
```bash
#!/bin/bash
DATABASE_FILE="/opt/app/data/production.db"
BACKUP_DIR="/opt/app/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Create backup
sqlite3 $DATABASE_FILE ".backup $BACKUP_DIR/backup_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/backup_$DATE.db"

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.db.gz" -mtime +30 -delete

echo "Backup completed: backup_$DATE.db.gz"
```

### Migration Management

```bash
# Run migrations manually
/opt/app/server -migrate

# Check migration status
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db status

# Rollback last migration (if needed)
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db down
```

## Monitoring and Logging

### Log Management

**With systemd:**
```bash
# View logs
sudo journalctl -u go-web-server -f

# Configure log rotation
sudo mkdir -p /etc/systemd/journald.conf.d
echo -e "[Journal]\nSystemMaxUse=1G\nMaxRetentionSec=30day" | sudo tee /etc/systemd/journald.conf.d/go-web-server.conf
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

### Health Checks

**External monitoring:**
```bash
# Simple health check
curl -f http://localhost:8080/health || exit 1

# Detailed monitoring script
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

## Scaling Considerations

### Horizontal Scaling

1. **Load Balancer Setup:**
   - Use Nginx, HAProxy, or cloud load balancer
   - Configure health checks
   - Enable session sticky if needed

2. **Database Considerations:**
   - SQLite is suitable for moderate loads
   - Consider PostgreSQL for high-traffic scenarios
   - Implement read replicas if needed

3. **Static Asset Delivery:**
   - Use CDN for static assets
   - Configure proper cache headers
   - Consider asset optimization

### Vertical Scaling

- **Memory:** Monitor memory usage and adjust limits
- **CPU:** Profile application for CPU bottlenecks  
- **Storage:** Monitor database file size and I/O
- **Network:** Configure connection limits appropriately

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check service status
sudo systemctl status go-web-server

# Check logs
sudo journalctl -u go-web-server -n 50

# Check file permissions
ls -la /opt/app/
```

**Database issues:**
```bash
# Check database file permissions
ls -la /opt/app/data/

# Verify database integrity
sqlite3 /opt/app/data/production.db "PRAGMA integrity_check;"

# Check migration status
goose -dir internal/store/migrations sqlite3 /opt/app/data/production.db status
```

**High memory usage:**
```bash
# Monitor memory usage
ps aux | grep server

# Check for memory leaks
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### Performance Optimization

1. **Database Optimization:**
   ```sql
   -- Analyze query performance
   EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = ?;
   
   -- Update statistics
   PRAGMA optimize;
   ```

2. **Application Profiling:**
   ```bash
   # Enable pprof in production (temporarily)
   export FEATURES_ENABLE_PPROF=true
   
   # Profile CPU usage
   curl http://localhost:8080/debug/pprof/profile > cpu.prof
   go tool pprof cpu.prof
   ```

3. **Connection Tuning:**
   - Adjust `DATABASE_MAX_CONNECTIONS`
   - Configure appropriate timeouts
   - Monitor connection pool usage

## Security Hardening

### System Level

- **User permissions:** Run as non-root user
- **File permissions:** Restrict access to binary and data
- **Network:** Use firewall to limit port access
- **Updates:** Keep system and dependencies updated

### Application Level

- **Environment:** Always use `production` environment
- **Logging:** Don't log sensitive information
- **Headers:** Configure security headers properly
- **CORS:** Restrict origins to your domains only

### Infrastructure Level

- **SSL/TLS:** Use strong cipher suites
- **WAF:** Consider Web Application Firewall
- **DDoS:** Implement rate limiting and DDoS protection
- **Monitoring:** Set up intrusion detection

This deployment guide ensures your Modern Go Web Server runs reliably and securely in production environments.