package bar

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Bar is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Bar cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.ClientErr(404, "Bar not found")
)
