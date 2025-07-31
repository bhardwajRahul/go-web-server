# Documentation

Complete documentation for the Modern Go Web Server.

## Quick Reference

- **[Development Guide](./development.md)** - Local setup and development workflow
- **[API Reference](./api.md)** - HTTP endpoints and HTMX integration
- **[Architecture](./architecture.md)** - System design and components
- **[Security](./security.md)** - CSRF, sanitization, and security headers
- **[Deployment](./deployment.md)** - Production deployment and configuration

## Getting Started

1. **Quick Start**: See the main [README.md](../README.md)
2. **Development**: Follow the [Development Guide](./development.md)
3. **Production**: Use the [Deployment Guide](./deployment.md)

## Stack Overview

- **Backend**: Echo v4 + SQLC + SQLite
- **Frontend**: Templ + HTMX 2.x + Pico.css
- **Security**: CSRF protection + input sanitization
- **Build**: Mage automation + Air hot reload
- **Deploy**: Single binary (~11MB) with zero dependencies
