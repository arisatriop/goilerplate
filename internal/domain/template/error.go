package template

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Template is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Template cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.ClientErr(404, "Template not found")
)
