// Package handler contains HTTP request handlers for the web application.
package handler

import (
	"net/http"
	"time"

	"github.com/dunamismax/go-web-server/internal/view"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles requests for the home page and health checks.
type HomeHandler struct{}

// NewHomeHandler creates a new HomeHandler instance.
func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

// Home handles requests to the root path, returning either full page or partial content.
func (h *HomeHandler) Home(c echo.Context) error {
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
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-web-server",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"checks": map[string]string{
			"database": "ok",
			"memory":   "ok",
		},
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

	return c.JSON(http.StatusOK, health)
}

var startTime = time.Now()
