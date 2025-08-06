package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// CSRFConfig defines the configuration for CSRF protection.
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes
	TokenLength int
	// TokenLookup defines where to look for the CSRF token
	// Format: "<source>:<name>"
	// Possible values:
	// - "header:<name>" (default: "header:X-CSRF-Token")
	// - "form:<name>" (default: "form:csrf_token")
	// - "query:<name>"
	TokenLookup string
	// CookieName is the name of the CSRF cookie
	CookieName string
	// CookieDomain is the domain of the CSRF cookie
	CookieDomain string
	// CookiePath is the path of the CSRF cookie
	CookiePath string
	// CookieSecure indicates if CSRF cookie is secure
	CookieSecure bool
	// CookieHTTPOnly indicates if CSRF cookie is HTTP only
	CookieHTTPOnly bool
	// CookieSameSite indicates SameSite policy for CSRF cookie
	CookieSameSite http.SameSite
	// CookieMaxAge indicates the max age of CSRF cookie
	CookieMaxAge int
	// ContextKey is the key used to store the CSRF token in context
	ContextKey string
	// ErrorHandler defines a function which is executed for an invalid CSRF token
	ErrorHandler CSRFErrorHandler
}

// CSRFErrorHandler defines a function which is executed for an invalid CSRF token.
type CSRFErrorHandler func(error, echo.Context) error

// DefaultCSRFConfig is the default CSRF middleware config.
var DefaultCSRFConfig = CSRFConfig{
	TokenLength:    32,
	TokenLookup:    "header:X-CSRF-Token,form:csrf_token",
	CookieName:     "_csrf",
	CookiePath:     "/",
	CookieSecure:   false,
	CookieHTTPOnly: true,
	CookieSameSite: http.SameSiteStrictMode,
	CookieMaxAge:   86400, // 24 hours
	ContextKey:     "csrf",
	ErrorHandler:   nil,
}

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
func CSRF() echo.MiddlewareFunc {
	return CSRFWithConfig(DefaultCSRFConfig)
}

// CSRFWithConfig returns a CSRF middleware with config.
func CSRFWithConfig(config CSRFConfig) echo.MiddlewareFunc {
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFConfig.TokenLength
	}

	if config.TokenLookup == "" {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}

	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}

	if config.CookiePath == "" {
		config.CookiePath = DefaultCSRFConfig.CookiePath
	}

	if config.CookieSameSite == 0 {
		config.CookieSameSite = DefaultCSRFConfig.CookieSameSite
	}

	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}

	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(err error, c echo.Context) error {
			return ErrCSRF.WithContext(c).WithInternal(err)
		}
	}

	// Parse token lookup
	parts := strings.Split(config.TokenLookup, ",")
	extractors := make([]csrfTokenExtractor, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		extractor := createCSRFTokenExtractor(part)
		if extractor != nil {
			extractors = append(extractors, extractor)
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip CSRF for safe methods
			method := c.Request().Method
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				token := generateCSRFToken(config.TokenLength)
				setCSRFCookie(c, config, token)
				c.Set(config.ContextKey, token)
				RecordCSRFTokenGenerated()

				return next(c)
			}

			// Get token from cookie
			cookie, err := c.Cookie(config.CookieName)
			if err != nil {
				return config.ErrorHandler(errors.New("CSRF cookie not found"), c)
			}

			cookieToken := cookie.Value

			// Get token from request
			var requestToken string
			for _, extractor := range extractors {
				requestToken = extractor(c)
				if requestToken != "" {
					break
				}
			}

			if requestToken == "" {
				return config.ErrorHandler(errors.New("CSRF token not found in request"), c)
			}

			// Validate token
			if !validateCSRFToken(cookieToken, requestToken) {
				RecordCSRFValidationFailure()

				return config.ErrorHandler(errors.New("CSRF token mismatch"), c)
			}

			// Generate new token for next request
			newToken := generateCSRFToken(config.TokenLength)
			setCSRFCookie(c, config, newToken)
			c.Set(config.ContextKey, newToken)
			RecordCSRFTokenGenerated()

			return next(c)
		}
	}
}

// csrfTokenExtractor extracts CSRF token from different sources.
type csrfTokenExtractor func(echo.Context) string

// createCSRFTokenExtractor creates token extractor from lookup string.
func createCSRFTokenExtractor(lookup string) csrfTokenExtractor {
	parts := strings.SplitN(lookup, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	source, name := parts[0], parts[1]
	switch source {
	case "header":
		return func(c echo.Context) string {
			return c.Request().Header.Get(name)
		}
	case "form":
		return func(c echo.Context) string {
			return c.FormValue(name)
		}
	case "query":
		return func(c echo.Context) string {
			return c.QueryParam(name)
		}
	}

	return nil
}

// generateCSRFToken generates a random CSRF token.
func generateCSRFToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based token if crypto/rand fails
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(bytes)
}

// validateCSRFToken validates CSRF token using constant-time comparison.
func validateCSRFToken(cookieToken, requestToken string) bool {
	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) == 1
}

// setCSRFCookie sets the CSRF cookie.
func setCSRFCookie(c echo.Context, config CSRFConfig, token string) {
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    token,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		MaxAge:   config.CookieMaxAge,
		Secure:   config.CookieSecure,
		HttpOnly: config.CookieHTTPOnly,
		SameSite: config.CookieSameSite,
	}
	c.SetCookie(cookie)
}

// GetCSRFToken returns the CSRF token from context.
func GetCSRFToken(c echo.Context) string {
	token := c.Get("csrf")
	if token == nil {
		return ""
	}

	if tokenStr, ok := token.(string); ok {
		return tokenStr
	}

	return ""
}
