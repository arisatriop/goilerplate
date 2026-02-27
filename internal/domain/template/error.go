package template

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.Error(409, "Code already exists")
	ErrAlreadyDeleted    = utils.Error(410, "Template is already deleted")
	ErrCannotBeDeleted   = utils.Error(403, "Template cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.Error(404, "Template not found")
)
