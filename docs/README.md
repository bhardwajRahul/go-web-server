# Documentation

Complete documentation for the Modern Go Stack web server.

## Quick Reference

- **[Development Guide](./development.md)** - Local setup, hot reload, database management, and daily workflow
- **[API Reference](./api.md)** - HTTP endpoints, HTMX integration, and CSRF protection
- **[Architecture](./architecture.md)** - System design, components, and technology decisions
- **[Security Guide](./security.md)** - CSRF, sanitization, headers, rate limiting, and monitoring
- **[Deployment Guide](./deployment.md)** - Production deployment with Caddy, Ubuntu, and Cloudflare

## Getting Started

1. **Quick Start**: See the main [README.md](../README.md) for immediate setup
2. **Development**: Follow the [Development Guide](./development.md) for local development
3. **Production**: Use the [Deployment Guide](./deployment.md) for Ubuntu + Caddy + Cloudflare

## The Modern Go Stack

### Core Technologies

- **Backend**: Echo v4 + SQLC v1.29.0 + SQLite (Pure Go)
- **Frontend**: Templ v0.3.850 + HTMX 2.0.6 + Pico.css v2
- **Security**: Custom CSRF + Input sanitization + Security headers
- **Build**: Mage automation + Air hot reload + Go 1.24+
- **Deploy**: Single binary (~11MB) + Zero dependencies

### Key Benefits

- **Type Safety**: SQLC for database, Templ for templates
- **Developer Experience**: Hot reload, comprehensive tooling, quality checks
- **Production Ready**: Enterprise security, structured errors, graceful shutdown
- **Performance**: Embedded assets, compiled templates, efficient middleware
- **Simplicity**: Single binary deployment, zero external dependencies

## Documentation Structure

Each guide is self-contained but references others where helpful:

### [Development Guide](./development.md)

- Quick setup and daily workflow
- Database development with SQLC and Goose migrations
- Template development with Templ and HTMX
- Handler patterns and error handling
- Configuration management
- Quality checks and debugging

### [API Reference](./api.md)

- Complete HTTP endpoint documentation
- HTMX integration patterns and custom events
- CSRF protection implementation
- Security features and rate limiting
- Response formats and error handling

### [Architecture](./architecture.md)

- System design and component structure
- Request flow and middleware stack (15 layers)
- Security architecture and defense-in-depth
- Database and configuration architecture
- Build system and asset embedding
- Performance characteristics

### [Security Guide](./security.md)

- Comprehensive security implementation
- CSRF protection with token rotation
- Input sanitization (XSS, SQL injection)
- Security headers and CORS configuration
- Rate limiting and error handling security
- Security monitoring and intrusion detection

### [Deployment Guide](./deployment.md)

- Production deployment for Ubuntu servers
- Caddy reverse proxy with automatic HTTPS
- Cloudflare integration for CDN and security
- Systemd service configuration and monitoring
- Database backup and maintenance
- Security hardening and performance tuning

## Stack Philosophy

The Modern Go Stack embodies Go's core principles:

1. **Simplicity over Complexity** - Clear, readable code and straightforward patterns
2. **Composition over Inheritance** - Small, focused components that work together
3. **Explicit over Implicit** - Clear error handling and explicit dependencies
4. **Performance by Design** - Compiled templates, embedded assets, efficient middleware
5. **Production First** - Every decision optimizes for deployment and operations

## Development Workflow

```bash
# 1. Setup (one time)
mage setup                 # Install tools and dependencies

# 2. Daily development
mage dev                   # Start with hot reload
mage generate              # Regenerate code (SQLC + Templ)
mage quality               # Run all quality checks

# 3. Production build
mage ci                    # Complete CI pipeline
mage build                 # Create production binary
```

## Support and Contributing

This template demonstrates production-ready patterns for the Modern Go Stack. Each component is carefully chosen and integrated to work seamlessly together while maintaining Go's core values of simplicity and performance.

For questions about specific implementation details, refer to the relevant guide above or examine the well-documented source code.
