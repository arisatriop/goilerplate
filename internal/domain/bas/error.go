package bas

import "goilerplate/pkg/utils"

var (
	ErrCodeAlreadyExists = utils.ClientErr(409, "Code already exists")
	ErrAlreadyDeleted    = utils.ClientErr(410, "Bas is already deleted")
	ErrCannotBeDeleted   = utils.ClientErr(403, "Bas cannot be deleted due to business rules")
	ErrNotFound          = utils.ClientErr(404, "Bas not found")
)
