package response

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NewValidator returns the validator every handler should use. It reports a field by the name
// the client actually sent — its json, query or params tag — so a validation error names
// "newPassword" rather than the Go field "NewPassword", or the lowercased "newpassword" that
// matched neither.
func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(clientFieldName)
	return validate
}

// clientFieldName picks the first wire-level tag set on the field. A tag of "-" means the field
// is not part of the request, and the validator skips it.
func clientFieldName(field reflect.StructField) string {
	for _, tag := range []string{"json", "query", "params", "form"} {
		name, _, _ := strings.Cut(field.Tag.Get(tag), ",")
		if name == "-" {
			return "-"
		}
		if name != "" {
			return name
		}
	}
	return field.Name
}
