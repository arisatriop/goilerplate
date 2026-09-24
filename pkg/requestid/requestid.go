// Package requestid decides which request ID a request is logged and answered under.
//
// A caller may send its own ID so a request can be followed across services. That value is
// attacker-controlled, and it is copied into every log line of the request and echoed in the
// response, so it is accepted only when it is short and made of characters that cannot break
// out of a log field or a header. Anything else is replaced, not rejected: a bad correlation ID
// is no reason to fail the request.
package requestid

import "github.com/google/uuid"

// MaxLength bounds an accepted ID. 128 leaves room for any common format — UUIDs, ULIDs,
// W3C trace IDs, prefixed IDs — without letting one request inflate every log line it writes.
const MaxLength = 128

// Resolve returns incoming when it is a valid request ID, otherwise a freshly generated one.
func Resolve(incoming string) string {
	if Valid(incoming) {
		return incoming
	}
	return New()
}

// New returns a new request ID: a UUIDv7, so IDs sort by the time the request arrived.
func New() string {
	return uuid.Must(uuid.NewV7()).String()
}

// Valid reports whether id is 1 to MaxLength characters of [A-Za-z0-9._-].
func Valid(id string) bool {
	if id == "" || len(id) > MaxLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		switch c := id[i]; {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '.', c == '_', c == '-':
		default:
			return false
		}
	}
	return true
}
