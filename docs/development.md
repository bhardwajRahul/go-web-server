# Development Guide

Local development setup and workflow for the Modern Go Stack.

## Quick Setup

```bash
# Clone and setup
git clone https://github.com/dunamismax/go-web-server.git
cd go-web-server

# Create your environment file
cp .env.example .env
# Edit .env with your database credentials (DATABASE_USER, DATABASE_PASSWORD, etc.)

# Install tools and dependencies
mage setup

# Ensure PostgreSQL is running locally
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Start development server with hot reload
mage dev
```

**Prerequisites:** Go 1.25+, PostgreSQL (local installation), Node.js + npm (for Tailwind CSS)

Server runs at `http://localhost:8080`

## Daily Commands

```bash
mage dev          # Start with hot reload
mage generate     # Generate SQLC + Templ code
mage quality      # Run vet + lint + vulncheck
mage build        # Production binary
mage ci           # Complete CI pipeline
```

## Database Development

**Local PostgreSQL:**

```bash
sudo systemctl start postgresql  # Start PostgreSQL service
mage migrate                     # Run migrations up
mage migrateDown                 # Rollback last migration
mage migrateStatus               # Show migration status
```

**Database Reset:**

```bash
mage reset        # Reset to fresh state with sample data
```

This command:

- Cleans build artifacts
- Removes generated code
- Regenerates code and templates
- Runs fresh migrations with sample data

## Environment Configuration

**Required Environment Variables (.env file):**

```bash
# Database Configuration
DATABASE_USER=your_user
DATABASE_PASSWORD=your_password
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=gowebserver
DATABASE_SSLMODE=disable

# Optional: Complete DATABASE_URL (overrides individual vars)
# DATABASE_URL=postgres://user:password@localhost:5432/gowebserver?sslmode=disable

# Authentication Configuration  
AUTH_JWT_SECRET=your-secret-key-change-in-production
AUTH_TOKEN_DURATION=24h
AUTH_REFRESH_DURATION=168h  # 7 days
AUTH_COOKIE_NAME=auth_token
AUTH_COOKIE_SECURE=true

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=
DEBUG=true
ENVIRONMENT=development
LOG_LEVEL=debug
LOG_FORMAT=text

# Feature Flags
FEATURES_ENABLE_METRICS=false
FEATURES_ENABLE_PPROF=false

# Security
SECURITY_ENABLE_CORS=true
SECURITY_ALLOWED_ORIGINS=*
```

## Code Generation Workflow

The application uses code generation for type safety:

```bash
# Generate database code from SQL
mage generateSqlc    # or: sqlc generate

# Generate templates from .templ files  
mage generateTempl   # or: templ generate

# Build Tailwind CSS
mage buildCSS       # or: npm run build-css

# Generate all
mage generate
```

**When to regenerate:**

- After modifying `internal/store/queries.sql`
- After modifying `internal/store/schema.sql`
- After creating/modifying `.templ` files
- After changing Tailwind CSS classes in templates
- After pulling changes that affect generated code

## Hot Reload Development

Using Air for hot reload:

```bash
mage dev  # Starts Air with automatic recompilation
```

**What triggers reload:**

- Go source file changes
- Template file changes (`.templ`)
- Tailwind CSS input file changes
- Static asset changes
- Configuration changes

**Air Configuration:**

- Builds to `tmp/main` for faster startups
- Excludes generated files from watch
- Includes SQL and template files in watch

## Quality Assurance

**Static Analysis:**

```bash
mage vet          # Go vet analysis
mage lint         # golangci-lint (comprehensive)
mage vulncheck    # Security vulnerability scanning
mage quality      # All quality checks
```

**Formatting:**

```bash
mage fmt          # Format with goimports + go mod tidy
```

**Complete CI Pipeline:**

```bash
mage ci           # generate + fmt + quality + build + info
```

## Testing Guidelines

**Manual Testing:**

1. **Authentication Flow**: Test login/register/logout with session management
2. **User Management**: Create, update, deactivate users
3. **HTMX Interactions**: Test dynamic page updates
4. **Error Handling**: Test validation and error responses
5. **CSRF Protection**: Test form submissions
6. **Theme Switching**: Test DaisyUI theme switching

**Browser Testing:**

- Chrome/Firefox/Safari compatibility
- Mobile responsiveness
- HTMX request/response inspection
- Network tab for partial updates
- Console for JavaScript errors

## Development Workflow

**Daily Development:**

1. Start PostgreSQL: `sudo systemctl start postgresql`
2. Start dev server: `mage dev`
3. Make changes to code
4. Auto-reload happens via Air
5. Test changes in browser
6. Run quality checks: `mage quality`
7. Commit changes with meaningful messages

**Adding New Features:**

1. **Database Changes**:
   - Update `internal/store/schema.sql` 
   - Run `mage migrate` to apply Atlas migrations
   - Update `queries.sql` if needed
   - Run `mage generate` to update Go code

2. **Handler Changes**:
   - Add routes in `internal/handler/routes.go`
   - Implement handlers in appropriate files
   - Add middleware if needed

3. **Template Changes**:
   - Create/modify `.templ` files
   - Update Tailwind CSS classes as needed
   - Run `mage generate` to compile templates and build CSS

4. **Testing**:
   - Manual testing in browser
   - Run `mage quality` for static analysis
   - Test CSRF protection on forms
   - Test HTMX interactions

## Common Issues

**Database Connection:**

```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check database exists
psql -U postgres -l

# Create database if missing
sudo -u postgres createdb gowebserver
sudo -u postgres createuser -P gowebserver
```

**Generation Issues:**

```bash
# Clean and regenerate everything
mage clean
mage generate
```

**Port Already in Use:**

```bash
# Kill process on port 8080
sudo lsof -ti:8080 | xargs kill -9
```

**Permission Issues:**

```bash
# Fix PostgreSQL authentication
sudo -u postgres psql
\password postgres
```

## IDE Configuration

**VS Code Extensions:**

- Go extension for Go development
- Templ extension for template syntax highlighting
- Tailwind CSS IntelliSense for CSS classes
- PostgreSQL extension for database management
- Thunder Client for API testing

**GoLand/IntelliJ:**

- Go plugin
- Database tools and SQL plugin
- File watchers for auto-generation

This development setup provides hot reload, comprehensive tooling, and immediate feedback for productive Go web development.
