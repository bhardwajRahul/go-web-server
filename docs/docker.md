# Docker Deployment Guide

Complete guide for deploying the Go Web Server using Docker and Docker Compose.

## Overview

The application uses a multi-service Docker architecture:

- **PostgreSQL**: Enterprise-grade database
- **Go Web Server**: The main application
- **Caddy**: Reverse proxy with automatic HTTPS

## Quick Start

```bash
# Clone repository
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server

# Create your environment file
cp .env.example .env
# Edit .env with your database credentials (DATABASE_USER, DATABASE_PASSWORD, etc.)

# Start all services
docker compose up --build

# Access application
open http://localhost        # Via Caddy reverse proxy
open http://localhost:8080   # Direct application access
```

## Service Configuration

### PostgreSQL Database

- **Image**: `postgres:16-alpine`
- **Port**: 5432 (exposed in development)
- **Credentials**: Configured via .env file environment variables
- **Volume**: `postgres_data` for data persistence
- **Migrations**: Auto-applied on container startup

### Go Web Server

- **Build**: Multi-stage Dockerfile for optimized image
- **Port**: 8080 (internal), exposed via Caddy
- **Health Check**: `/health` endpoint monitoring
- **Environment**: Production-ready configuration
- **Security**: Non-root user, minimal attack surface

### Caddy Reverse Proxy

- **Image**: `caddy:2-alpine`
- **Ports**: 80 (HTTP), 443 (HTTPS)
- **Features**: Automatic HTTPS, HTTP/2, compression
- **Configuration**: `Caddyfile` with security headers

## Production Deployment

### 1. Environment Setup

Create production environment file:

```bash
# .env.production
POSTGRES_PASSWORD=your_secure_password
DATABASE_URL=postgres://user:your_secure_password@postgres:5432/gowebserver?sslmode=disable

# Domain configuration (update Caddyfile)
DOMAIN=yourdomain.com
```

### 2. Caddyfile Configuration

Update `Caddyfile` for your domain:

```caddyfile
yourdomain.com, www.yourdomain.com {
    # Automatic HTTPS with Let's Encrypt
    reverse_proxy app:8080 {
        health_uri /health
        health_interval 30s
        health_timeout 10s
    }
    
    encode gzip zstd
    
    header {
        -Server
        Strict-Transport-Security "max-age=63072000; includeSubDomains; preload"
    }
}
```

### 3. Production Deployment

```bash
# Production deployment
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Or using environment file
docker compose --env-file .env.production up -d
```

## Development Workflow

### Local Development

```bash
# Create your environment file
cp .env.example .env
# Edit .env with your database credentials

# Start database only
docker compose up postgres -d

# Use local Go development
mage setup
mage dev
```

### Full Docker Development

```bash
# Start all services with hot reload
docker compose up --build

# View logs
docker compose logs -f app

# Access database
docker exec -it gowebserver-postgres psql -U user -d gowebserver
```

## Management Commands

### Database Management

```bash
# Run migrations
mage migrate

# Database backup
docker exec gowebserver-postgres pg_dump -U user gowebserver > backup.sql

# Database restore
docker exec -i gowebserver-postgres psql -U user gowebserver < backup.sql

# Connect to database
docker exec -it gowebserver-postgres psql -U user -d gowebserver
```

### Application Management

```bash
# View application logs
docker compose logs -f app

# Restart application
docker compose restart app

# Update application
git pull
docker compose up --build -d app

# Scale application (load balancing)
docker compose up -d --scale app=3
```

### System Monitoring

```bash
# Service status
docker compose ps

# Resource usage
docker stats

# Health checks
curl http://localhost/health
curl http://localhost:8080/metrics
```

## Docker Commands Reference

### Development

```bash
mage docker           # Start all services
mage dockerDown       # Stop all services
mage dockerReset      # Reset environment (remove volumes)
mage dockerLogs       # Show all service logs
```

### Manual Operations

```bash
# Build application image
docker build -t gowebserver .

# Run application manually
docker run -p 8080:8080 --env-file .env gowebserver

# Network inspection
docker network ls
docker network inspect gowebserver-network
```

## Troubleshooting

### Common Issues

**Application won't start:**

```bash
# Check dependencies
docker compose ps
docker compose logs postgres

# Verify database connection
docker exec gowebserver-postgres pg_isready -U user
```

**Caddy SSL issues:**

```bash
# Check domain DNS
nslookup yourdomain.com

# Verify Caddy configuration
docker compose logs caddy

# Test HTTP before HTTPS
curl -I http://yourdomain.com
```

**Database connection issues:**

```bash
# Check PostgreSQL logs
docker compose logs postgres

# Test connection
docker exec -it gowebserver-postgres psql -U user -d gowebserver -c "SELECT 1;"
```

### Performance Optimization

**Database tuning in docker-compose.yml:**

```yaml
postgres:
  environment:
    POSTGRES_SHARED_BUFFERS: 256MB
    POSTGRES_EFFECTIVE_CACHE_SIZE: 1GB
    POSTGRES_MAX_CONNECTIONS: 100
```

**Application scaling:**

```bash
docker compose up -d --scale app=3
```

## Security Considerations

### Production Security

1. **Change default passwords**
2. **Use secrets management**
3. **Enable container scanning**
4. **Regular security updates**
5. **Network segmentation**

### Environment Variables

```bash
# Required for production
POSTGRES_PASSWORD=secure_password
DATABASE_URL=postgres://user:secure_password@postgres:5432/gowebserver?sslmode=disable
ENVIRONMENT=production

# Optional security features  
SECURITY_ENABLE_CORS=false
FEATURES_ENABLE_RATE_LIMITING=true
FEATURES_ENABLE_CSRF=true
```

## Monitoring and Observability

### Metrics Collection

- Prometheus metrics at `/metrics`
- Health check at `/health`
- Structured logging with JSON output

### Log Management

```bash
# Centralized logging
docker compose logs --follow --tail=100

# Service-specific logs
docker compose logs app --follow
docker compose logs postgres --follow
docker compose logs caddy --follow
```

## Backup and Recovery

### Automated Backups

Create backup script:

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
docker exec gowebserver-postgres pg_dump -U user gowebserver | gzip > "backup_${DATE}.sql.gz"

# Keep only last 7 days
find . -name "backup_*.sql.gz" -mtime +7 -delete
```

### Recovery Process

```bash
# Stop services
docker compose down

# Remove old data
docker volume rm gowebserver_postgres_data

# Start services
docker compose up -d postgres

# Wait for database to be ready
sleep 30

# Restore data
gunzip -c backup_20240101_120000.sql.gz | docker exec -i gowebserver-postgres psql -U user gowebserver

# Start application
docker compose up -d app caddy
```

This Docker-first architecture provides a robust, scalable, and secure foundation for deploying the Go Web Server in any environment.
