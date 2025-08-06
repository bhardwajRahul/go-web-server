# Documentation

Complete documentation for the Modern Go Stack web server.

## Quick Reference

- **[Development Guide](./development.md)** - Local setup and workflow
- **[API Reference](./api.md)** - HTTP endpoints and HTMX integration
- **[Architecture](./architecture.md)** - System design and components
- **[Security Guide](./security.md)** - Security implementation
- **[Deployment Guide](./deployment.md)** - Production deployment

## The Modern Go Stack

**Core Technologies:**

- Backend: Echo v4 + SQLC + SQLite (Pure Go)
- Frontend: Templ v0.3.924 + HTMX 2.0.6 + Pico.css v2
- Build: Mage + Air hot reload + Go 1.24+
- Deploy: Single binary (~11MB), zero dependencies

**Key Benefits:**

- Type-safe database queries and templates
- Hot reload development with comprehensive tooling
- Production-ready security and error handling
- Single binary deployment with embedded assets

## Quick Start

```bash
# Setup and run
mage setup && mage dev

# Quality checks and build
mage quality && mage build
```

## Stack Philosophy

1. **Simplicity** - Clear, readable patterns
2. **Performance** - Compiled templates, embedded assets
3. **Production First** - Built for deployment and operations
4. **Type Safety** - SQLC + Templ eliminate runtime errors

This template demonstrates production-ready patterns while maintaining Go's simplicity and performance characteristics.
