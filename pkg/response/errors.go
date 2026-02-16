package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorDetail represents a single validation error
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator errors to a standardized format
func FormatValidationErrors(err error) []ValidationErrorDetail {
	var details []ValidationErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			detail := ValidationErrorDetail{
				Field: strings.ToLower(fieldError.Field()),
				Tag:   fieldError.Tag(),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}

			// Custom error messages based on validation tag
			switch fieldError.Tag() {
			case "required":
				detail.Message = fmt.Sprintf("%s is required", detail.Field)
			case "email":
				detail.Message = fmt.Sprintf("%s must be a valid email address", detail.Field)
			case "min":
				detail.Message = fmt.Sprintf("%s must be at least %s characters", detail.Field, fieldError.Param())
			case "max":
				detail.Message = fmt.Sprintf("%s must not exceed %s characters", detail.Field, fieldError.Param())
			case "len":
				detail.Message = fmt.Sprintf("%s must be exactly %s characters", detail.Field, fieldError.Param())
			case "gt":
				detail.Message = fmt.Sprintf("%s must be greater than %s", detail.Field, fieldError.Param())
			case "gte":
				detail.Message = fmt.Sprintf("%s must be greater than or equal to %s", detail.Field, fieldError.Param())
			case "lt":
				detail.Message = fmt.Sprintf("%s must be less than %s", detail.Field, fieldError.Param())
			case "lte":
				detail.Message = fmt.Sprintf("%s must be less than or equal to %s", detail.Field, fieldError.Param())
			case "uuid":
				detail.Message = fmt.Sprintf("%s must be a valid UUID", detail.Field)
			case "url":
				detail.Message = fmt.Sprintf("%s must be a valid URL", detail.Field)
			case "alpha":
				detail.Message = fmt.Sprintf("%s must contain only alphabetic characters", detail.Field)
			case "alphanum":
				detail.Message = fmt.Sprintf("%s must contain only alphanumeric characters", detail.Field)
			case "numeric":
				detail.Message = fmt.Sprintf("%s must be numeric", detail.Field)
			case "json":
				detail.Message = fmt.Sprintf("%s must be valid JSON", detail.Field)
			case "oneof":
				detail.Message = fmt.Sprintf("%s must be one of: %s", detail.Field, fieldError.Param())
			default:
				detail.Message = fmt.Sprintf("%s is invalid", detail.Field)
			}

			details = append(details, detail)
		}
	}

	return details
}
