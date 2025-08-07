<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/go-logo.png" alt="Go Web Server Template Logo" width="400" />
</p>

<p align="center">
  <a href="https://github.com/dunamismax/go-web-server">
    <img src="https://readme-typing-svg.demolab.com/?font=Fira+Code&size=24&pause=1000&color=00ADD8&center=true&vCenter=true&width=900&lines=The+Modern+Go+Stack;Echo+v4+Framework+with+Type-Safe+Templates;HTMX+Dynamic+UX+without+JavaScript;SQLC+Generated+Queries+with+Pure+Go+SQLite;CSRF+Protection+and+Input+Sanitization;Structured+Error+Handling+and+Request+Tracing;Hot+Reload+Development+with+Mage+Automation;Single+Binary+Deployment+at+14MB;Production-Ready+Security+Middleware;Zero+External+Dependencies" alt="Typing SVG" />
  </a>
</p>

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8.svg?logo=go" alt="Go Version"></a>
  <a href="https://echo.labstack.com/"><img src="https://img.shields.io/badge/Framework-Echo_v4-00ADD8.svg?logo=go" alt="Echo Framework"></a>
  <a href="https://templ.guide/"><img src="https://img.shields.io/badge/Templates-Templ-00ADD8.svg?logo=go" alt="Templ"></a>
  <a href="https://htmx.org/"><img src="https://img.shields.io/badge/Frontend-HTMX_2.x-3D72D7.svg?logo=htmx" alt="HTMX"></a>
  <a href="https://picocss.com/"><img src="https://img.shields.io/badge/CSS-Pico.css_v2-13795B.svg" alt="Pico.css"></a>
  <a href="https://sqlc.dev/"><img src="https://img.shields.io/badge/Queries-SQLC-00ADD8.svg?logo=go" alt="SQLC"></a>
  <a href="https://www.sqlite.org/"><img src="https://img.shields.io/badge/Database-SQLite-003B57.svg?logo=sqlite" alt="SQLite"></a>
  <a href="https://pkg.go.dev/modernc.org/sqlite"><img src="https://img.shields.io/badge/Driver-Pure_Go-00ADD8.svg?logo=go" alt="Pure Go SQLite"></a>
  <a href="https://pkg.go.dev/log/slog"><img src="https://img.shields.io/badge/Logging-slog-00ADD8.svg?logo=go" alt="Go slog"></a>
  <a href="https://github.com/knadh/koanf"><img src="https://img.shields.io/badge/Config-Koanf-00ADD8.svg?logo=go" alt="Koanf"></a>
  <a href="https://github.com/pressly/goose"><img src="https://img.shields.io/badge/Migrations-Goose-00ADD8.svg?logo=go" alt="Goose"></a>
  <a href="https://magefile.org/"><img src="https://img.shields.io/badge/Build-Mage-purple.svg?logo=go" alt="Mage"></a>
  <a href="https://github.com/air-verse/air"><img src="https://img.shields.io/badge/HotReload-Air-FF6B6B.svg?logo=go" alt="Air"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License"></a>
</p>

---

## About

A production-ready template for modern web applications using **The Modern Go Stack** - a cohesive technology stack for building high-performance, maintainable applications. Creates single, self-contained binaries with zero external dependencies.

**Key Features:**

- **Echo v4 + Templ + HTMX**: High-performance web framework with type-safe templates and dynamic UX
- **SQLC + SQLite + Pure Go Driver**: Type-safe database operations with zero CGO dependencies
- **Prometheus Metrics**: Comprehensive observability and performance monitoring
- **Enterprise Security**: CSRF protection, input sanitization, structured error handling, request tracing
- **Mage Build System**: Go-based automation with comprehensive quality checks
- **Production Ready**: Rate limiting, CORS, secure headers, graceful shutdown
- **Developer Experience**: Hot reload with Air, database migrations with Goose, multi-source config

## Tech Stack

| Layer          | Technology                                                  | Purpose                                |
| -------------- | ----------------------------------------------------------- | -------------------------------------- |
| **Language**   | [Go 1.24+](https://go.dev/doc/)                             | Latest performance & language features |
| **Framework**  | [Echo v4](https://echo.labstack.com/)                       | High-performance web framework         |
| **Templates**  | [Templ](https://templ.guide/)                      | Type-safe Go HTML components           |
| **Frontend**   | [HTMX](https://htmx.org/)                             | Dynamic interactions with smooth UX    |
| **CSS**        | [Pico.css v2](https://picocss.com/)                         | Semantic CSS with dark/light themes    |
| **Logging**    | [slog](https://pkg.go.dev/log/slog)                         | Structured logging with JSON output    |
| **Database**   | [SQLite](https://www.sqlite.org/)                           | Self-contained, serverless database    |
| **Queries**    | [SQLC](https://sqlc.dev/)                           | Generate type-safe Go from SQL         |
| **Metrics**    | [Prometheus](https://prometheus.io/)                        | Performance monitoring & observability |
| **DB Driver**  | [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure Go, CGO-free SQLite driver        |
| **Assets**     | [Go Embed](https://pkg.go.dev/embed)                        | Single binary with embedded resources  |
| **Config**     | [Koanf](https://github.com/knadh/koanf)                     | Multi-source configuration management  |
| **Migrations** | [Goose](https://github.com/pressly/goose)                   | Database migration management          |
| **Build**      | [Mage](https://magefile.org/)                               | Go-based build automation              |
| **Hot Reload** | [Air](https://github.com/air-verse/air)                     | Development server with live reload    |

---

## Quick Start

```bash
# Clone and setup
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server
go mod tidy

# Install development tools and dependencies
mage setup

# Start development server with hot reload
mage dev

# Or build and run production binary
mage run

# Server starts at http://localhost:8080
```

**Requirements:**

- Go 1.24+
- Mage build tool (`go install github.com/magefile/mage@latest`)

**Note:** First run of `mage setup` installs all development tools automatically.

## Documentation

**[Complete Documentation](docs/)** - Comprehensive guides for development, deployment, security, and architecture.

| Guide | Description |
|-------|-------------|
| **[Development Guide](docs/development.md)** | Local setup, hot reload, database management, and daily workflow |
| **[API Reference](docs/api.md)** | HTTP endpoints, HTMX integration, and CSRF protection |
| **[Architecture](docs/architecture.md)** | System design, components, and technology decisions |
| **[Security Guide](docs/security.md)** | CSRF, sanitization, headers, rate limiting, and monitoring |
| **[Deployment Guide](docs/deployment.md)** | Production deployment, configuration, and scaling |

---

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/gopher-mage.svg" alt="Gopher Mage" width="150" />
</p>

## Mage Commands

Run `mage help` to see all available commands and their aliases.

**Development:**

```bash
mage setup (s)        # Install tools and dependencies
mage generate (g)     # Generate sqlc and templ code
mage dev (d)          # Start development server with hot reload
mage run (r)          # Build and run server
mage build (b)        # Build production binary
```

**Database:**

```bash
mage migrate (m)      # Run database migrations up
mage migrateDown      # Roll back last migration
mage migrateStatus    # Show migration status
```

**Quality & Production:**

```bash
mage fmt (f)          # Format code with goimports and tidy modules
mage vet (v)          # Run go vet static analysis
mage lint (l)         # Run golangci-lint comprehensive linting
mage vulncheck (vc)   # Check for security vulnerabilities
mage quality (q)      # Run all quality checks
mage ci               # Complete CI pipeline
mage clean (c)        # Clean build artifacts
```

**Observability & Monitoring:**

```bash
# Enable Prometheus metrics (via environment variables)
FEATURES_ENABLE_METRICS=true mage run
# Then access metrics at: http://localhost:8080/metrics

# Enhanced health checks with database connectivity
curl http://localhost:8080/health
```

## Live Demo

### Web Application (`localhost:8080`)

Interactive user management application demonstrating:

- **CRUD Operations**: Type-safe database queries with CSRF protection
- **Real-time Updates**: HTMX interactions with smooth page transitions
- **Responsive Design**: Automatic dark/light theme switching with Pico.css
- **Enterprise Security**: Input sanitization and structured error handling

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/go-web-server-user-screenshot.png" alt="Go Web Server Screenshot" width="800" />
</p>

> **Easter Egg**: The default user database comes pre-populated with Robert Griesemer, Rob Pike, and Ken Thompson - the three brilliant minds who created the Go programming language at Google starting in 2007. A small tribute to the creators of the language that powers this entire stack!

## Project Structure

```sh
go-web-server/
├── cmd/web/              # Application entry point
├── docs/                 # Complete documentation
├── internal/
│   ├── config/           # Koanf configuration management
│   ├── handler/          # HTTP handlers with Echo routes
│   ├── middleware/       # Security, validation, error handling
│   ├── store/            # Database layer with SQLC
│   │   └── migrations/   # Goose database migrations
│   ├── ui/               # Static assets (embedded)
│   └── view/             # Templ templates and components
├── bin/                  # Compiled binaries
├── magefile.go          # Mage build automation
├── .golangci.yml        # Linter configuration
└── sqlc.yaml            # SQLC configuration

```

---

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/gopher-aviator.jpg" alt="Go Gopher" width="400" />
</p>

## Single Binary Deployment

```bash
mage build  # Creates optimized binary in bin/server (~14MB)
```

The binary includes embedded Pico.css, HTMX, Templ templates, and SQLite database. **Zero external dependencies**, single file deployment with instant startup.

## Key Features Demonstrated

**Modern Web Stack:**

- Echo framework with comprehensive middleware stack
- Type-safe Templ templates with reusable components
- HTMX dynamic interactions with smooth page transitions
- Pico.css semantic styling with automatic dark/light themes
- SQLC type-safe database queries with pure Go SQLite driver
- Structured logging with slog and configurable JSON output
- Prometheus metrics for observability and performance monitoring

**Developer Experience:**

- Hot reloading with Air for rapid development
- Comprehensive error handling with structured logging
- Static analysis suite (golangci-lint, govulncheck, go vet)
- Mage build automation with goimports and templ formatting
- Single-command CI pipeline with quality checks

**Production Ready:**

- Enterprise security with CSRF protection and input sanitization
- Structured error handling with request tracing and monitoring
- Multi-source configuration with Koanf (JSON, YAML, ENV)
- Database migrations with Goose and graceful shutdown
- Single binary deployment (~14MB) with embedded assets
- Zero external dependencies and CGO-free compilation

---

<p align="center">
  <a href="https://buymeacoffee.com/dunamismax" target="_blank">
    <img src="https://github.com/dunamismax/images/blob/main/golang/buy-coffee-go.gif" alt="Buy Me A Coffee" style="height: 150px !important;" />
  </a>
</p>

<p align="center">
  <a href="https://twitter.com/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Twitter-%231DA1F2.svg?&style=for-the-badge&logo=twitter&logoColor=white" alt="Twitter"></a>
  <a href="https://bsky.app/profile/dunamismax.bsky.social" target="_blank"><img src="https://img.shields.io/badge/Bluesky-blue?style=for-the-badge&logo=bluesky&logoColor=white" alt="Bluesky"></a>
  <a href="https://reddit.com/user/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Reddit-%23FF4500.svg?&style=for-the-badge&logo=reddit&logoColor=white" alt="Reddit"></a>
  <a href="https://discord.com/users/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Discord-dunamismax-7289DA.svg?style=for-the-badge&logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://signal.me/#p/+dunamismax.66" target="_blank"><img src="https://img.shields.io/badge/Signal-dunamismax.66-3A76F0.svg?style=for-the-badge&logo=signal&logoColor=white" alt="Signal"></a>
</p>

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>The Modern Go Stack</strong><br>
  <sub>Echo • Templ • HTMX • Pico.css • SQLC • SQLite • slog • Koanf • Goose • Mage • Air</sub>
</p>

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/gopher-running-jumping.gif" alt="Gopher Running and Jumping" width="600" />
</p>

---

"The "Modern Go Stack" is a powerful and elegant solution that aligns beautifully with Go's core principles. It is an excellent starting point for many new projects, and any decision to deviate from it should be driven by specific, demanding requirements." - Me

---
