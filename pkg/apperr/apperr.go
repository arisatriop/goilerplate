// Package apperr describes errors a client caused and can act on, without committing to a
// transport.
//
// The domain says what went wrong — the resource does not exist, the input is invalid, the
// caller is not who they claim — and each delivery layer decides how that is spelled: HTTP as a
// status code (pkg/response), gRPC as a status code with an ErrorInfo detail (pkg/grpcresponse).
// The domain therefore never imports net/http, and a new transport needs one mapping table rather
// than a translation of HTTP codes.
//
// Every Error carries a Code: a stable, machine-readable, snake_case identifier such as
// "bar_not_found". Clients branch on the code; the Message is prose for humans and may change.
package apperr

import "errors"

// Kind is the category of a client error. It decides the transport status and nothing else.
type Kind uint8

// The kinds a client can cause. Anything that is not one of these is a server fault and is
// reported as a generic internal error.
const (
	// Invalid: the request is malformed or breaks a rule about its own content.
	Invalid Kind = iota + 1
	// Unauthenticated: no credential, or one that is invalid, expired or revoked.
	Unauthenticated
	// Forbidden: the caller is known but not allowed to do this.
	Forbidden
	// NotFound: the resource does not exist — or exists but is not the caller's to see.
	NotFound
	// Conflict: the request is valid but clashes with the current state (a duplicate key, say).
	Conflict
)

// Error is a client error. Compare with errors.Is against a sentinel declared with New: two
// Errors match when their Kind and Code match, so a sentinel still matches after WithCause.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	cause   error
}

// New declares a client error. Message must be safe to show to the caller.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

// Error returns the client-facing message.
func (e *Error) Error() string { return e.Message }

// Unwrap returns the underlying cause, if any, for logging and errors.Is/As.
func (e *Error) Unwrap() error { return e.cause }

// Is reports whether target is an *Error of the same Kind and Code.
func (e *Error) Is(target error) bool {
	var other *Error
	if !errors.As(target, &other) {
		return false
	}
	return e.Kind == other.Kind && e.Code == other.Code
}

// WithCause returns a copy of e that wraps cause, so the reason can be logged while the client
// still sees only e.Message.
func (e *Error) WithCause(cause error) *Error {
	copied := *e
	copied.cause = cause
	return &copied
}

// As returns the first *Error in err's chain.
func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
