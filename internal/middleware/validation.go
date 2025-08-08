package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator interface for custom validation.
type CustomValidator interface {
	Validate() error
}

// ValidationError represents a validation error with field information.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
	Tag     string `json:"tag,omitempty"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors is a slice of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

// Global validator instance
var validate = validator.New()

func init() {
	// Register custom validations
	registerCustomValidations()

	// Use JSON tags for field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return strings.ToLower(fld.Name)
		}
		return name
	})
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations() {
	// Register password validation
	err := validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return len(password) >= 8 &&
			strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
			strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
			strings.ContainsAny(password, "0123456789")
	})
	if err != nil {
		panic("failed to register password validation: " + err.Error())
	}
}

// ValidateStruct validates a struct using go-playground/validator.
func ValidateStruct(s interface{}) ValidationErrors {
	var validationErrors ValidationErrors

	err := validate.Struct(s)
	if err == nil {
		return validationErrors
	}

	// Handle validation errors
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fieldErr := range ve {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fieldErr.Field(),
				Message: getErrorMessage(fieldErr),
				Value:   fieldErr.Value(),
				Tag:     fieldErr.Tag(),
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "invalid email format"
	case "url":
		return "invalid URL format"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("minimum length is %s", fe.Param())
		}
		return fmt.Sprintf("minimum value is %s", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("maximum length is %s", fe.Param())
		}
		return fmt.Sprintf("maximum value is %s", fe.Param())
	case "len":
		return fmt.Sprintf("length must be %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "password":
		return "password must be at least 8 characters with uppercase, lowercase, and numeric characters"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "alpha":
		return "must contain only alphabetic characters"
	case "numeric":
		return "must contain only numeric characters"
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "datetime":
		return "must be a valid datetime"
	default:
		return fmt.Sprintf("validation failed for tag '%s'", fe.Tag())
	}
}

// ValidateAndBind is an Echo middleware that validates request body.
func ValidateAndBind(target interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create new instance of target type
			targetType := reflect.TypeOf(target)
			if targetType.Kind() == reflect.Ptr {
				targetType = targetType.Elem()
			}

			instance := reflect.New(targetType).Interface()

			// Bind request data
			if err := c.Bind(instance); err != nil {
				slog.Error("failed to bind request data", "error", err)

				return NewAppError(
					ErrorTypeValidation,
					http.StatusBadRequest,
					"Invalid request format",
				).WithContext(c).WithInternal(err)
			}

			// Run custom validation if implemented
			if customValidator, ok := instance.(CustomValidator); ok {
				if err := customValidator.Validate(); err != nil {
					slog.Warn("custom validation failed", "error", err)

					return NewAppError(
						ErrorTypeValidation,
						http.StatusBadRequest,
						"Custom validation failed",
					).WithContext(c).WithInternal(err)
				}
			}

			// Run struct validation using go-playground/validator
			if validationErrors := ValidateStruct(instance); len(validationErrors) > 0 {
				slog.Warn("struct validation failed", "errors", validationErrors)

				return NewAppErrorWithDetails(
					ErrorTypeValidation,
					http.StatusBadRequest,
					"Validation failed",
					validationErrors,
				).WithContext(c)
			}

			// Store validated data in context
			c.Set("validated", instance)

			return next(c)
		}
	}
}

// GetValidated retrieves validated data from context.
func GetValidated[T any](c echo.Context) (*T, error) {
	validated := c.Get("validated")
	if validated == nil {
		return nil, errors.New("no validated data found in context")
	}

	data, ok := validated.(*T)
	if !ok {
		return nil, errors.New("validated data is not of expected type")
	}

	return data, nil
}

// Validate validates a single struct instance
func Validate(s interface{}) error {
	if validationErrors := ValidateStruct(s); len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}
