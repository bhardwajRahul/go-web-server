package middleware

import (
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// SanitizeConfig defines the configuration for input sanitization.
type SanitizeConfig struct {
	// SanitizeHTML enables HTML sanitization
	SanitizeHTML bool
	// SanitizeSQL enables basic SQL injection protection
	SanitizeSQL bool
	// SanitizeXSS enables XSS protection
	SanitizeXSS bool
	// CustomSanitizers allows custom sanitization functions
	CustomSanitizers []func(string) string
}

// DefaultSanitizeConfig is the default sanitization config.
var DefaultSanitizeConfig = SanitizeConfig{
	SanitizeHTML: true,
	SanitizeSQL:  true,
	SanitizeXSS:  true,
}

// Sanitize returns input sanitization middleware.
func Sanitize() echo.MiddlewareFunc {
	return SanitizeWithConfig(DefaultSanitizeConfig)
}

// SanitizeWithConfig returns input sanitization middleware with config.
func SanitizeWithConfig(config SanitizeConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a sanitizing request wrapper
			req := &sanitizingRequest{
				Request: c.Request(),
				config:  config,
			}
			c.SetRequest(req.Request)

			return next(c)
		}
	}
}

// sanitizingRequest wraps http.Request to sanitize form values.
type sanitizingRequest struct {
	*http.Request

	config SanitizeConfig
}

// FormValue returns the sanitized form value for the provided key.
func (r *sanitizingRequest) FormValue(key string) string {
	value := r.Request.FormValue(key)

	return r.sanitizeValue(value)
}

// PostFormValue returns the sanitized POST form value for the provided key.
func (r *sanitizingRequest) PostFormValue(key string) string {
	value := r.Request.PostFormValue(key)

	return r.sanitizeValue(value)
}

// sanitizeValue applies all configured sanitization rules.
func (r *sanitizingRequest) sanitizeValue(value string) string {
	if value == "" {
		return value
	}

	result := value

	// Apply HTML sanitization
	if r.config.SanitizeHTML {
		result = sanitizeHTML(result)
	}

	// Apply XSS protection
	if r.config.SanitizeXSS {
		result = sanitizeXSS(result)
	}

	// Apply SQL injection protection
	if r.config.SanitizeSQL {
		result = sanitizeSQL(result)
	}

	// Apply custom sanitizers
	for _, sanitizer := range r.config.CustomSanitizers {
		result = sanitizer(result)
	}

	return result
}

// sanitizeHTML escapes HTML characters to prevent HTML injection.
func sanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// sanitizeXSS removes or escapes potential XSS vectors.
func sanitizeXSS(input string) string {
	// Remove or escape dangerous patterns
	dangerous := []string{
		"javascript:",
		"vbscript:",
		"data:",
		"blob:",
		"<script",
		"</script>",
		"<iframe",
		"</iframe>",
		"<object",
		"</object>",
		"<embed",
		"</embed>",
		"<form",
		"</form>",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
	}

	result := strings.ToLower(input)
	for _, pattern := range dangerous {
		result = strings.ReplaceAll(result, pattern, "")
	}

	// Remove any remaining event handlers
	eventHandlerRegex := regexp.MustCompile(`on\w+\s*=`)
	result = eventHandlerRegex.ReplaceAllString(result, "")

	// If the result is significantly different, return escaped version
	if len(result) < int(float64(len(input))*0.8) {
		return html.EscapeString(input)
	}

	return input
}

// sanitizeSQL provides basic SQL injection protection.
func sanitizeSQL(input string) string {
	// Remove SQL comment patterns
	sqlComments := []string{
		"--",
		"/*",
		"*/",
		"#",
	}

	result := input
	for _, comment := range sqlComments {
		result = strings.ReplaceAll(result, comment, "")
	}

	// Remove dangerous SQL keywords (case-insensitive)
	dangerousPatterns := []string{
		"union select",
		"union all select",
		"drop table",
		"drop database",
		"delete from",
		"truncate table",
		"alter table",
		"create table",
		"insert into",
		"update set",
		"exec(",
		"execute(",
		"sp_",
		"xp_",
	}

	lowerResult := strings.ToLower(result)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerResult, pattern) {
			// If dangerous pattern found, escape the entire string
			return strings.ReplaceAll(input, "'", "''")
		}
	}

	// Basic quote escaping
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

// SanitizeString provides a utility function to sanitize individual strings.
func SanitizeString(input string, config SanitizeConfig) string {
	sanitizer := &sanitizingRequest{config: config}

	return sanitizer.sanitizeValue(input)
}

// Common sanitization presets.
var (
	// HTMLSanitizeConfig sanitizes HTML content.
	HTMLSanitizeConfig = SanitizeConfig{
		SanitizeHTML: true,
		SanitizeXSS:  true,
		SanitizeSQL:  false,
	}

	// FormSanitizeConfig sanitizes form inputs.
	FormSanitizeConfig = SanitizeConfig{
		SanitizeHTML: true,
		SanitizeXSS:  true,
		SanitizeSQL:  true,
	}

	// SQLSanitizeConfig focuses on SQL injection prevention.
	SQLSanitizeConfig = SanitizeConfig{
		SanitizeHTML: false,
		SanitizeXSS:  false,
		SanitizeSQL:  true,
	}
)
