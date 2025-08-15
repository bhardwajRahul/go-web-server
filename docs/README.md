# Documentation

Complete documentation for the Modern Go Stack web server template.

## Quick Reference

- **[Development Guide](./development.md)** - Local setup, hot reload, and daily workflow
- **[API Reference](./api.md)** - HTTP endpoints, HTMX integration, and CSRF protection
- **[Architecture](./architecture.md)** - System design, components, and technology decisions
- **[Security Guide](./security.md)** - CSRF, sanitization, headers, and monitoring
- **[Deployment Guide](./deployment.md)** - Production deployment configuration
- **[Ubuntu Deployment Guide](./ubuntu-deployment.md)** - Complete Ubuntu SystemD deployment

## The Modern Go Stack

A production-ready web application template using a cohesive technology stack for building high-performance, maintainable applications.

**Core Technologies:**

- **Backend**: Go 1.25+ + Echo v4 + SQLC + PostgreSQL (pgx/v5 driver)
- **Frontend**: Templ v0.3.924 + HTMX 2.x + Tailwind CSS + DaisyUI
- **Security**: Session Authentication + CSRF Protection + Input Sanitization
- **Build**: Mage automation + Air hot reload + comprehensive quality checks
- **Config**: Koanf multi-source + structured logging (slog) + request tracing
- **Deploy**: Single binary (~15MB) + Ubuntu SystemD + embedded assets

**Key Benefits:**

- **Type Safety**: SQLC generates type-safe database code, Templ provides compile-time template validation
- **Security First**: Multi-layer defense with CSRF, input sanitization, security headers, and rate limiting
- **Developer Experience**: Hot reload, comprehensive tooling, static analysis, vulnerability scanning
- **Production Ready**: Single binary deployment, graceful shutdown, connection pooling, structured logging
- **High Performance**: Compiled templates, embedded assets, PostgreSQL with pgx driver
- **Enterprise Features**: Session authentication, Atlas migrations, comprehensive error handling

## Quick Start

```bash
# Clone and setup
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server

# Create environment configuration
cp .env.example .env
# Edit .env with your database credentials

# Install tools and dependencies
mage setup

# Ensure PostgreSQL is running
sudo systemctl start postgresql

# Start development server with hot reload
mage dev  # Runs at http://localhost:8080
```

## Quality & Production

```bash
# Comprehensive quality checks
mage quality     # vet + lint + vulncheck
mage ci          # Complete CI pipeline

# Production build
mage build       # Creates bin/server (~15MB)

# Database management
mage migrate     # Run migrations
mage reset       # Reset to fresh state
```

## Stack Philosophy

1. **Simplicity** - Clear, maintainable patterns that align with Go's philosophy
2. **Performance** - Compiled templates, connection pooling, embedded assets, minimal allocations
3. **Security** - Defense-in-depth with CSRF, sanitization, headers, and structured error handling
4. **Type Safety** - SQLC + Templ eliminate runtime errors through compile-time validation
5. **Production First** - Built for deployment, monitoring, and operations from day one
6. **Developer Experience** - Hot reload, comprehensive tooling, and immediate feedback

## Features Demonstrated

**Modern Web Development:**

- Server-side rendering with dynamic HTMX interactions
- Type-safe HTML templates with Templ
- Progressive enhancement without JavaScript complexity
- Modern Tailwind CSS + DaisyUI styling with multiple themes
- Real-time updates with smooth page transitions

**Enterprise Security:**

- Session authentication with Argon2id password hashing
- CSRF protection with token rotation
- Input sanitization for XSS/SQL injection prevention
- Security headers (CSP, HSTS, X-Frame-Options)
- Rate limiting and structured error handling

**Production Operations:**

- Structured logging with request tracing
- Database connection pooling and health checks
- Atlas declarative schema management
- Graceful shutdown and configuration management
- Single binary deployment with SystemD integration

This template provides an excellent starting point for Go web applications, demonstrating production-ready patterns while maintaining simplicity and developer productivity.
