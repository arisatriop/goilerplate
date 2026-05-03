package bazs

import "goilerplate/pkg/utils"

var (
	// Business logic errors
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Bazs is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Bazs cannot be deleted due to business rules")

	// Operation errors
	ErrNotFound = utils.ClientErr(404, "Bazs not found")
)
