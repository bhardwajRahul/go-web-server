// Package handler contains HTTP request handlers for the web application.
package handler

import (
	"net/http"
	"time"

	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/dunamismax/go-web-server/internal/view"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles requests for the home page and health checks.
type HomeHandler struct {
	store *store.Store
}

// NewHomeHandler creates a new HomeHandler instance.
func NewHomeHandler(s *store.Store) *HomeHandler {
	return &HomeHandler{store: s}
}

// Home handles requests to the root path, returning either full page or partial content.
func (h *HomeHandler) Home(c echo.Context) error {
	// Set CSRF token in response header for initial requests
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		component := view.HomeContent()
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Return full page with layout
	component := view.Home()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Demo provides a demonstration of HTMX functionality
func (h *HomeHandler) Demo(c echo.Context) error {
	demoData := struct {
		Message    string
		Features   []string
		ServerTime string
		RequestID  string
	}{
		Message:    "🎉 Demo successful! This content was loaded dynamically using HTMX.",
		Features:   []string{"Server-side rendering", "Dynamic content loading", "No page refresh", "Smooth animations"},
		ServerTime: time.Now().Format("3:04:05 PM MST"),
		RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
	}

	// Check if this is an HTMX request for formatted HTML display
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		component := view.DemoContent(demoData.Message, demoData.Features, demoData.ServerTime, demoData.RequestID)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Set response headers for JSON response
	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSON(http.StatusOK, demoData)
}

// Health provides a comprehensive health check endpoint
func (h *HomeHandler) Health(c echo.Context) error {
	ctx := c.Request().Context()
	checks := make(map[string]string)
	overallStatus := "ok"

	// Database connectivity check
	if h.store != nil {
		if _, err := h.store.CountUsers(ctx); err != nil {
			checks["database"] = "error"
			overallStatus = "degraded"
		} else {
			checks["database"] = "ok"
		}

		// Database connection stats
		if db := h.store.DB(); db != nil {
			if stats := db.Stats(); stats.OpenConnections > 0 {
				checks["database_connections"] = "ok"
			} else {
				checks["database_connections"] = "warning"
				if overallStatus == "ok" {
					overallStatus = "warning"
				}
			}
		}
	} else {
		checks["database"] = "error"
		overallStatus = "error"
	}

	// Memory check (basic)
	checks["memory"] = "ok"

	health := map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-web-server",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"checks":    checks,
	}

	// Check if this is an HTMX request for formatted HTML display
	if c.Request().Header.Get("HX-Request") == "true" {
		component := view.HealthCheck(
			health["status"].(string),
			health["service"].(string),
			health["version"].(string),
			health["uptime"].(string),
			health["timestamp"].(string),
			health["checks"].(map[string]string),
		)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Set response headers for JSON response
	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Set appropriate HTTP status based on health
	var statusCode int
	switch overallStatus {
	case "error":
		statusCode = http.StatusServiceUnavailable
	case "degraded", "warning":
		statusCode = http.StatusPartialContent
	default:
		statusCode = http.StatusOK
	}

	return c.JSON(statusCode, health)
}

var startTime = time.Now()
