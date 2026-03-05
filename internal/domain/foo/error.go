package foo

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Foo is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Foo cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.ClientErr(404, "Foo not found")
)
