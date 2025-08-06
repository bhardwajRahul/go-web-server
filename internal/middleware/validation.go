package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Validator interface for custom validation.
type Validator interface {
	Validate() error
}

// ValidationError represents a validation error with field information.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
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

// ValidateStruct validates a struct using reflection and validate tags.
func ValidateStruct(s any) ValidationErrors {
	var errors ValidationErrors

	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		tagValue := fieldType.Tag.Get("validate")
		if tagValue == "" {
			continue
		}

		fieldName := getFieldName(fieldType)
		fieldValue := field.Interface()

		// Parse validation rules
		rules := parseValidationRules(tagValue)

		for _, rule := range rules {
			if err := validateRule(fieldName, fieldValue, rule); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

// ValidationRule represents a single validation rule.
type ValidationRule struct {
	Name  string
	Param string
}

// parseValidationRules parses validation tag into rules.
func parseValidationRules(tag string) []ValidationRule {
	var rules []ValidationRule

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules = append(rules, ValidationRule{
				Name:  strings.TrimSpace(kv[0]),
				Param: strings.TrimSpace(kv[1]),
			})
		} else {
			rules = append(rules, ValidationRule{
				Name: part,
			})
		}
	}

	return rules
}

// validateRule validates a single rule against a field value.
func validateRule(fieldName string, value any, rule ValidationRule) *ValidationError {
	switch rule.Name {
	case "required":
		return validateRequired(fieldName, value)
	case "min":
		return validateMin(fieldName, value, rule.Param)
	case "max":
		return validateMax(fieldName, value, rule.Param)
	case "email":
		return validateEmail(fieldName, value)
	case "url":
		return validateURL(fieldName, value)
	case "oneof":
		return validateOneOf(fieldName, value, rule.Param)
	}

	return nil
}

// validateRequired checks if a field is required.
func validateRequired(fieldName string, value any) *ValidationError {
	if isZeroValue(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: "field is required",
			Value:   value,
		}
	}

	return nil
}

// validateMin validates minimum length/value.
func validateMin(fieldName string, value any, param string) *ValidationError {
	minVal, err := strconv.Atoi(param)
	if err != nil {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid min parameter",
		}
	}

	switch v := value.(type) {
	case string:
		if len(v) < minVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("minimum length is %d", minVal),
				Value:   len(v),
			}
		}
	case int, int8, int16, int32, int64:
		val := reflect.ValueOf(v).Int()
		if val < int64(minVal) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("minimum value is %d", minVal),
				Value:   val,
			}
		}
	}

	return nil
}

// validateMax validates maximum length/value.
func validateMax(fieldName string, value any, param string) *ValidationError {
	maxVal, err := strconv.Atoi(param)
	if err != nil {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid max parameter",
		}
	}

	switch v := value.(type) {
	case string:
		if len(v) > maxVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("maximum length is %d", maxVal),
				Value:   len(v),
			}
		}
	case int, int8, int16, int32, int64:
		val := reflect.ValueOf(v).Int()
		if val > int64(maxVal) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("maximum value is %d", maxVal),
				Value:   val,
			}
		}
	}

	return nil
}

// validateEmail validates email format.
func validateEmail(fieldName string, value any) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:   fieldName,
			Message: "field must be a string",
		}
	}

	if !strings.Contains(str, "@") || !strings.Contains(str, ".") {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid email format",
			Value:   str,
		}
	}

	return nil
}

// validateURL validates URL format.
func validateURL(fieldName string, value any) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:   fieldName,
			Message: "field must be a string",
		}
	}

	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid URL format",
			Value:   str,
		}
	}

	return nil
}

// validateOneOf validates that value is one of allowed values.
func validateOneOf(fieldName string, value any, param string) *ValidationError {
	allowedValues := strings.Split(param, " ")
	valueStr := fmt.Sprintf("%v", value)

	for _, allowed := range allowedValues {
		if valueStr == allowed {
			return nil
		}
	}

	return &ValidationError{
		Field:   fieldName,
		Message: "value must be one of: " + param,
		Value:   value,
	}
}

// isZeroValue checks if a value is zero/empty.
func isZeroValue(value any) bool {
	if value == nil {
		return true
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	}

	return false
}

// getFieldName gets the field name for validation (prefers json tag).
func getFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	return strings.ToLower(field.Name)
}

// ValidateAndBind is an Echo middleware that validates request body.
func ValidateAndBind(target any) echo.MiddlewareFunc {
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

				return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
					"error":   "invalid request format",
					"details": err.Error(),
				})
			}

			// Run custom validation if implemented
			if validator, ok := instance.(Validator); ok {
				if err := validator.Validate(); err != nil {
					slog.Warn("custom validation failed", "error", err)

					return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
						"error":   "validation failed",
						"details": err.Error(),
					})
				}
			}

			// Run struct validation
			if validationErrors := ValidateStruct(instance); len(validationErrors) > 0 {
				slog.Warn("struct validation failed", "errors", validationErrors)

				return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
					"error":  "validation failed",
					"fields": validationErrors,
				})
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
