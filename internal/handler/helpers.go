package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/labstack/echo/v4"
)

// isHtmxRequest checks if the request is an HTMX request
func isHtmxRequest(c echo.Context) bool {
	return c.Request().Header.Get(HtmxRequest) == HtmxRequestHeader
}

// htmxRedirect sets the HX-Redirect header and returns a JSON response
func htmxRedirect(c echo.Context, url, message string) error {
	c.Response().Header().Set(HtmxRedirect, url)
	return c.JSON(http.StatusOK, map[string]string{
		"message": message,
	})
}

// redirectOrHtmx handles both regular redirects and HTMX redirects
func redirectOrHtmx(c echo.Context, url, message string) error {
	if isHtmxRequest(c) {
		return htmxRedirect(c, url, message)
	}
	return c.Redirect(http.StatusFound, url)
}

// isJSONRequest checks if the request accepts JSON
func isJSONRequest(c echo.Context) bool {
	return c.Request().Header.Get("Accept") == ContentTypeJSON
}

// setupCSRFHeaders sets CSRF token in response headers if available
func setupCSRFHeaders(c echo.Context) string {
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}
	return token
}

// renderWithCSRF renders content with CSRF handling for both HTMX and regular requests
func renderWithCSRF(c echo.Context, htmxComponent, fullPageComponent, basicComponent templ.Component) error {
	setupCSRFHeaders(c)

	if isHtmxRequest(c) {
		return htmxComponent.Render(c.Request().Context(), c.Response().Writer)
	}

	// Try to use the full page component with CSRF first
	if fullPageComponent != nil {
		return fullPageComponent.Render(c.Request().Context(), c.Response().Writer)
	}

	// Fallback to basic component
	return basicComponent.Render(c.Request().Context(), c.Response().Writer)
}

// Error helpers for common error patterns

// validationError creates a validation error with context
func validationError(c echo.Context, message string, err error) error {
	return middleware.NewAppError(
		middleware.ErrorTypeValidation,
		http.StatusBadRequest,
		message,
	).WithContext(c).WithInternal(err)
}

// validationErrorWithDetails creates a validation error with validation details
func validationErrorWithDetails(c echo.Context, message string, details interface{}) error {
	return middleware.NewAppErrorWithDetails(
		middleware.ErrorTypeValidation,
		http.StatusBadRequest,
		message,
		details,
	).WithContext(c)
}

// authenticationError creates an authentication error
func authenticationError(c echo.Context, message string) error {
	return middleware.NewAppError(
		middleware.ErrorTypeAuthentication,
		http.StatusUnauthorized,
		message,
	).WithContext(c)
}

// internalError creates an internal server error with context
func internalError(c echo.Context, message string, err error) error {
	return middleware.NewAppError(
		middleware.ErrorTypeInternal,
		http.StatusInternalServerError,
		message,
	).WithContext(c).WithInternal(err)
}

// notFoundError creates a not found error
func notFoundError(c echo.Context, message string) error {
	return middleware.NewAppError(
		middleware.ErrorTypeNotFound,
		http.StatusNotFound,
		message,
	).WithContext(c)
}

// parseIDParam parses and validates an ID parameter from the URL
func parseIDParam(c echo.Context, paramName string) (int64, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid ID format",
		).WithContext(c).WithInternal(err)
	}
	return id, nil
}

// logAndReturnError logs an error and returns an app error
func logAndReturnError(c echo.Context, operation string, err error, statusCode int, userMessage string) error {
	slog.Error("Operation failed",
		"operation", operation,
		"error", err,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	return middleware.NewAppError(
		middleware.ErrorTypeInternal,
		statusCode,
		userMessage,
	).WithContext(c).WithInternal(err)
}

// stringPtr returns a pointer to string if not empty, nil otherwise
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
