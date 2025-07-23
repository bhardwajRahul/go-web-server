// Package main provides the entry point for the Go web server application.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dunamismax/go-web-server/internal/config"
	"github.com/dunamismax/go-web-server/internal/handler"
	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:generate go install github.com/a-h/templ/cmd/templ@latest
//go:generate go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
//go:generate templ generate
//go:generate sh -c "cd ../../ && sqlc generate"

func main() {
	// Load configuration
	cfg := config.New()

	// Setup structured logging
	var logger *slog.Logger
	if cfg.App.LogFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.GetLogLevel(),
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.GetLogLevel(),
		}))
	}
	slog.SetDefault(logger)

	slog.Info("Starting Go Web Server",
		"version", "1.0.0",
		"environment", cfg.App.Environment,
		"go_version", "1.24+",
		"port", cfg.Server.Port,
		"debug", cfg.App.Debug)

	// Initialize database store
	store, err := store.NewStore(cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err, "database_url", cfg.Database.URL)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Run migrations if enabled
	if cfg.Database.RunMigrations {
		slog.Info("Running database migrations with Goose")
		if err := runGooseMigrations(cfg.Database.URL); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	// Initialize schema (fallback if migrations not used)
	if err := store.InitSchema(); err != nil {
		slog.Error("failed to initialize schema", "error", err)
		os.Exit(1)
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.Debug = cfg.App.Debug

	// Configure custom error handler
	e.HTTPErrorHandler = middleware.ErrorHandler

	// Set custom 404 and 405 handlers
	e.RouteNotFound("/*", middleware.NotFoundHandler)
	e.Add("*", "/*", middleware.MethodNotAllowedHandler)

	// Configure timeouts
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Middleware stack (order matters)

	// Custom recovery middleware (should be first)
	e.Use(middleware.RecoveryMiddleware())

	// Security headers middleware
	e.Use(middleware.SecurityHeadersMiddleware())

	// Validation error middleware
	e.Use(middleware.ValidationErrorMiddleware())

	// Timeout error middleware
	e.Use(middleware.TimeoutErrorHandler())

	// Request ID middleware for tracing
	e.Use(echomiddleware.RequestID())

	// Structured logging middleware
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		LogMethod:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogUserAgent: cfg.App.Debug,
		LogValuesFunc: func(_ echo.Context, v echomiddleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.Info("request",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency.String(),
					"remote_ip", v.RemoteIP,
					"request_id", v.RequestID)
			} else {
				slog.Error("request error",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency.String(),
					"remote_ip", v.RemoteIP,
					"request_id", v.RequestID,
					"error", v.Error)
			}
			return nil
		},
	}))

	// Security middleware
	e.Use(echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'",
	}))

	// CORS middleware
	if cfg.Security.EnableCORS {
		e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
			AllowOrigins: cfg.Security.AllowedOrigins,
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
			AllowHeaders: []string{"*"},
			MaxAge:       86400,
		}))
	}

	// Rate limiting
	e.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStore(20),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(_ echo.Context, err error) error {
			return middleware.ErrTooManyRequests.WithInternal(err)
		},
	}))

	// Timeout middleware
	e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		Timeout: cfg.Server.ReadTimeout,
	}))

	// Add environment to context for error handling
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("environment", cfg.App.Environment)
			return next(c)
		}
	})

	// Initialize handlers and register routes
	handlers := handler.NewHandlers(store)
	if err := handler.RegisterRoutes(e, handlers); err != nil {
		slog.Error("failed to register routes", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		slog.Info("Server starting", "address", address)

		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown complete")
}

// runGooseMigrations runs database migrations using Goose
func runGooseMigrations(databaseURL string) error {
	// Open database connection
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Set Goose dialect to SQLite
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db, "internal/store/migrations"); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	slog.Info("Database migrations completed successfully with Goose")
	return nil
}
