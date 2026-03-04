package example

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Example is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Example cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.ClientErr(404, "Example not found")
)
