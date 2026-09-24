// This file is a TEMPLATE, not a feature.
//
// foo is the blank scaffold a new domain is copied from; bar next to it is the worked example
// with the bodies filled in. The panics are deliberate — they are what an unimplemented method
// should do in a template, because a silent zero value would let a half-copied domain look like
// it works. Its HTTP routes are not registered; see internal/delivery/http/router/public.go.
//
// The copy procedure is in .claude/skills/crud-operations/SKILL.md.

package foo

import "goilerplate/pkg/apperr"

// Declare each error the domain can return with a kind (which decides the status) and a stable
// snake_case code prefixed with the domain name. See bar/error.go for a worked example.
var (
	ErrNotFound          = apperr.New(apperr.NotFound, "foo_not_found", "Foo not found")
	ErrCodeAlreadyExists = apperr.New(apperr.Conflict, "foo_code_already_exists", "Code already exists")
)
