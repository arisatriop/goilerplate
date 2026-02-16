package zexample

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.Error(409, "code already exists")
	ErrAlreadyDeleted    = utils.Error(410, "example is already deleted")
	ErrCannotBeDeleted   = utils.Error(403, "example cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.Error(404, "example not found")
)
