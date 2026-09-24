package bar

import "goilerplate/pkg/apperr"

// Client errors of the bar domain. The delivery layer decides the status from the kind; the code
// is the stable identifier clients branch on.
var (
	ErrNotFound             = apperr.New(apperr.NotFound, "bar_not_found", "Bar not found")
	ErrCodeAlreadyExists    = apperr.New(apperr.Conflict, "bar_code_already_exists", "Code already exists")
	ErrDuplicateCodeInBatch = apperr.New(apperr.Invalid, "bar_code_repeated_in_request", "The same code appears more than once in the request")

	ErrCodeRequired        = apperr.New(apperr.Invalid, "bar_code_required", "code is required")
	ErrCodeFormat          = apperr.New(apperr.Invalid, "bar_code_invalid", "code must start with 'EXP'")
	ErrDescriptionRequired = apperr.New(apperr.Invalid, "bar_required", "bar is required")
)
