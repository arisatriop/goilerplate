package user

import "goilerplate/pkg/apperr"

// ErrEmailAlreadyRegistered: registration with an address that already has an account. Same kind
// and code as auth.ErrEmailAlreadyRegistered, so either matches the other with errors.Is.
var ErrEmailAlreadyRegistered = apperr.New(apperr.Conflict, "email_already_registered", "Email is already registered")
