// Package middleware provides custom middleware functions for the Echo web framework.
package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	// ErrorTypeValidation represents validation-related errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeAuthentication represents authentication-related errors
	ErrorTypeAuthentication ErrorType = "authentication"
	// ErrorTypeAuthorization represents authorization-related errors
	ErrorTypeAuthorization ErrorType = "authorization"
	// ErrorTypeNotFound represents resource not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeConflict represents resource conflict errors
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeRateLimit represents rate limiting errors
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeCSRF represents CSRF token errors
	ErrorTypeCSRF ErrorType = "csrf"
	// ErrorTypeSanitization represents input sanitization errors
	ErrorTypeSanitization ErrorType = "sanitization"
)

// AppError represents an application-specific error with enhanced context
type AppError struct {
	Type      ErrorType `json:"type"`
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Details   any       `json:"details,omitempty"`
	Internal  error     `json:"-"` // Internal error (not exposed to client)
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
	Path      string    `json:"path,omitempty"`
	Method    string    `json:"method,omitempty"`
}

func (e AppError) Error() string {
	if e.Internal != nil {
		return e.Internal.Error()
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code int, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(errorType ErrorType, code int, message string, details any) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WithContext adds request context to the error
func (e *AppError) WithContext(c echo.Context) *AppError {
	if c != nil {
		e.RequestID = c.Response().Header().Get(echo.HeaderXRequestID)
		e.Path = c.Request().URL.Path
		e.Method = c.Request().Method
	}
	return e
}

// WithInternal adds an internal error for logging purposes
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// Common application errors
var (
	ErrBadRequest         = NewAppError(ErrorTypeValidation, http.StatusBadRequest, "Bad request")
	ErrUnauthorized       = NewAppError(ErrorTypeAuthentication, http.StatusUnauthorized, "Unauthorized")
	ErrForbidden          = NewAppError(ErrorTypeAuthorization, http.StatusForbidden, "Forbidden")
	ErrNotFound           = NewAppError(ErrorTypeNotFound, http.StatusNotFound, "Resource not found")
	ErrConflict           = NewAppError(ErrorTypeConflict, http.StatusConflict, "Resource already exists")
	ErrTooManyRequests    = NewAppError(ErrorTypeRateLimit, http.StatusTooManyRequests, "Rate limit exceeded")
	ErrInternalServer     = NewAppError(ErrorTypeInternal, http.StatusInternalServerError, "Internal server error")
	ErrServiceUnavailable = NewAppError(ErrorTypeExternal, http.StatusServiceUnavailable, "Service unavailable")
	ErrTimeout            = NewAppError(ErrorTypeTimeout, http.StatusRequestTimeout, "Request timeout")
	ErrCSRF               = NewAppError(ErrorTypeCSRF, http.StatusForbidden, "Invalid CSRF token")
)

// ErrorResponse represents the JSON error response structure with enhanced metadata
type ErrorResponse struct {
	Type      ErrorType `json:"type"`
	Error     string    `json:"error"`
	Message   string    `json:"message,omitempty"`
	Details   any       `json:"details,omitempty"`
	Code      int       `json:"code"`
	Path      string    `json:"path,omitempty"`
	Method    string    `json:"method,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// ErrorHandler is a custom Echo error handler with enhanced error tracking
func ErrorHandler(err error, c echo.Context) {
	var (
		errorType = ErrorTypeInternal
		code      = http.StatusInternalServerError
		message   = "Internal server error"
		details   any
	)

	// Handle different error types
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Application error
		errorType = appErr.Type
		code = appErr.Code
		message = appErr.Message
		details = appErr.Details

		// Add context if not already present
		if appErr.RequestID == "" || appErr.Path == "" {
			appErr = appErr.WithContext(c)
		}

		// Log internal error if present
		if appErr.Internal != nil {
			slog.Error("application error",
				"type", appErr.Type,
				"error", appErr.Internal,
				"code", code,
				"message", message,
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
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
			"method", c.Request().Method,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
	} else {
		// Generic error
		slog.Error("unhandled error",
			"error", err,
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
			"user_agent", c.Request().UserAgent(),
			"remote_ip", c.RealIP())
	}

	// Don't send error response if response was already sent
	if c.Response().Committed {
		return
	}

	// Create enhanced error response
	errorResp := ErrorResponse{
		Type:      errorType,
		Error:     http.StatusText(code),
		Message:   message,
		Details:   details,
		Code:      code,
		Path:      c.Request().URL.Path,
		Method:    c.Request().Method,
		RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		Timestamp: fmt.Sprintf("%d", c.Request().Context().Value("timestamp")),
	}

	// Set timestamp if not available from context
	if errorResp.Timestamp == "<nil>" || errorResp.Timestamp == "" {
		errorResp.Timestamp = "server-time"
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
	return ErrNotFound.WithContext(c)
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler(c echo.Context) error {
	return NewAppError(ErrorTypeValidation, http.StatusMethodNotAllowed, "Method not allowed").WithContext(c)
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
					ErrorTypeValidation,
					http.StatusBadRequest,
					"Validation failed",
					validationErrs,
				).WithContext(c)
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
				return ErrTimeout.WithContext(c)
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
