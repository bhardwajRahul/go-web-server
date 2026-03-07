package middleware

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

const defaultMultipartMemory = 32 << 20

// SanitizeConfig defines the configuration for request input normalization.
type SanitizeConfig struct {
	// TrimSpace removes leading and trailing whitespace.
	TrimSpace bool
	// StripNullBytes removes NUL bytes from input values.
	StripNullBytes bool
	// CustomSanitizers allows additional caller-provided normalization.
	CustomSanitizers []func(string) string
}

// DefaultSanitizeConfig is the default request normalization config.
var DefaultSanitizeConfig = SanitizeConfig{
	TrimSpace:      true,
	StripNullBytes: true,
}

// Sanitize returns request normalization middleware.
func Sanitize() echo.MiddlewareFunc {
	return SanitizeWithConfig(DefaultSanitizeConfig)
}

// SanitizeWithConfig returns request normalization middleware with config.
func SanitizeWithConfig(config SanitizeConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().Clone(c.Request().Context())
			if err := normalizeRequest(req, config); err != nil {
				return NewAppError(
					ErrorTypeValidation,
					http.StatusBadRequest,
					"Invalid form payload",
				).WithContext(c).WithInternal(err)
			}
			c.SetRequest(req)

			return next(c)
		}
	}
}

func normalizeRequest(req *http.Request, config SanitizeConfig) error {
	if req == nil {
		return nil
	}

	queryValues := req.URL.Query()
	normalizeValues(queryValues, config)
	req.URL.RawQuery = queryValues.Encode()

	contentType := req.Header.Get(echo.HeaderContentType)
	if strings.HasPrefix(contentType, echo.MIMEMultipartForm) {
		if err := req.ParseMultipartForm(defaultMultipartMemory); err != nil {
			return err
		}
		normalizeMultipartForm(req.MultipartForm, config)
	} else if err := req.ParseForm(); err != nil {
		return err
	}

	normalizeValues(req.Form, config)
	normalizeValues(req.PostForm, config)

	return nil
}

func normalizeMultipartForm(form *multipart.Form, config SanitizeConfig) {
	if form == nil {
		return
	}

	normalizeValues(form.Value, config)
}

func normalizeValues(values url.Values, config SanitizeConfig) {
	if len(values) == 0 {
		return
	}

	for key, items := range values {
		for idx, item := range items {
			values[key][idx] = SanitizeString(item, config)
		}
	}
}

// SanitizeString normalizes an individual string value.
func SanitizeString(input string, config SanitizeConfig) string {
	if input == "" {
		return input
	}

	result := input
	if config.StripNullBytes {
		result = strings.ReplaceAll(result, "\x00", "")
	}
	if config.TrimSpace {
		result = strings.TrimSpace(result)
	}
	for _, sanitizer := range config.CustomSanitizers {
		result = sanitizer(result)
	}

	return result
}

// Common normalization presets.
var (
	// FormSanitizeConfig normalizes form inputs.
	FormSanitizeConfig = DefaultSanitizeConfig
)
