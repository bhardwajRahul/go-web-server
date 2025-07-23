package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AppError represents an application-specific error
type AppError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Details  any    `json:"details,omitempty"`
	Internal error  `json:"-"` // Internal error (not exposed to client)
}

func (e AppError) Error() string {
	if e.Internal != nil {
		return e.Internal.Error()
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code int, message string, details any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WithInternal adds an internal error for logging purposes
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// Common application errors
var (
	ErrBadRequest         = NewAppError(http.StatusBadRequest, "Bad request")
	ErrUnauthorized       = NewAppError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden          = NewAppError(http.StatusForbidden, "Forbidden")
	ErrNotFound           = NewAppError(http.StatusNotFound, "Resource not found")
	ErrConflict           = NewAppError(http.StatusConflict, "Resource already exists")
	ErrTooManyRequests    = NewAppError(http.StatusTooManyRequests, "Rate limit exceeded")
	ErrInternalServer     = NewAppError(http.StatusInternalServerError, "Internal server error")
	ErrServiceUnavailable = NewAppError(http.StatusServiceUnavailable, "Service unavailable")
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
	Code    int    `json:"code"`
	Path    string `json:"path,omitempty"`
	Method  string `json:"method,omitempty"`
}

// ErrorHandler is a custom Echo error handler
func ErrorHandler(err error, c echo.Context) {
	var (
		code    = http.StatusInternalServerError
		message = "Internal server error"
		details any
	)

	// Handle different error types
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Application error
		code = appErr.Code
		message = appErr.Message
		details = appErr.Details

		// Log internal error if present
		if appErr.Internal != nil {
			slog.Error("application error",
				"error", appErr.Internal,
				"code", code,
				"message", message,
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"user_agent", c.Request().UserAgent(),
				"remote_ip", c.RealIP())
		}
	} else if echoErr, ok := err.(*echo.HTTPError); ok {
		// Echo HTTP error
		code = echoErr.Code
		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		} else {
			message = http.StatusText(code)
		}
		details = echoErr.Internal

		slog.Warn("HTTP error",
			"error", err,
			"code", code,
			"message", message,
			"path", c.Request().URL.Path,
			"method", c.Request().Method)
	} else {
		// Generic error
		slog.Error("unhandled error",
			"error", err,
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
			"user_agent", c.Request().UserAgent(),
			"remote_ip", c.RealIP())
	}

	// Don't send error response if response was already sent
	if c.Response().Committed {
		return
	}

	// Create error response
	errorResp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Details: details,
		Code:    code,
		Path:    c.Request().URL.Path,
		Method:  c.Request().Method,
	}

	// Remove details in production for security
	if code >= 500 {
		errorResp.Details = nil
		if c.Get("environment") == "production" {
			errorResp.Message = "Internal server error"
		}
	}

	// Send JSON error response
	if err := c.JSON(code, errorResp); err != nil {
		slog.Error("failed to send error response", "error", err)
	}
}

// RecoveryMiddleware creates a custom recovery middleware
func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = errors.New("unknown panic")
					}

					slog.Error("panic recovered",
						"error", err,
						"panic", r,
						"path", c.Request().URL.Path,
						"method", c.Request().Method,
						"user_agent", c.Request().UserAgent(),
						"remote_ip", c.RealIP())

					// Create app error for panic
					appErr := ErrInternalServer.WithInternal(err)
					ErrorHandler(appErr, c)
				}
			}()

			return next(c)
		}
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(c echo.Context) error {
	return ErrNotFound
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler(c echo.Context) error {
	return NewAppError(http.StatusMethodNotAllowed, "Method not allowed")
}

// ValidationErrorMiddleware converts validation errors to app errors
func ValidationErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// Check if it's a validation error
			var validationErrs ValidationErrors
			if errors.As(err, &validationErrs) {
				return NewAppErrorWithDetails(
					http.StatusBadRequest,
					"Validation failed",
					validationErrs,
				)
			}

			return err
		}
	}
}

// TimeoutErrorHandler handles timeout errors
func TimeoutErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// Check if it's a timeout error by checking the error message
			if err.Error() == "timeout" || err.Error() == "request timeout" {
				return NewAppError(http.StatusRequestTimeout, "Request timeout")
			}

			return err
		}
	}
}

// SecurityHeadersMiddleware adds additional security headers not covered by Echo's secure middleware
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Additional security headers not covered by Echo's SecureMiddleware
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")

			return next(c)
		}
	}
}
