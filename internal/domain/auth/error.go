package auth

import "errors"

// ErrNotFound is returned by the repository when an update or delete matches no record.
var ErrNotFound = errors.New("auth: record not found")
