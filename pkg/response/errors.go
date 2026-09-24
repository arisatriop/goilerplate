package response

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorDetail describes one rejected field.
//
// It deliberately carries no copy of the submitted value. The value is what the client just
// sent, so echoing it tells them nothing new — and when the field is a password, echoing it put
// the password in the response body and, through the response logger, in the logs.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator errors to a standardized format. Field names are the
// client-facing names when the validator came from NewValidator.
func FormatValidationErrors(err error) []ValidationErrorDetail {
	details := []ValidationErrorDetail{}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			detail := ValidationErrorDetail{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
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
