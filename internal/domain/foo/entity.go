// This file is a TEMPLATE, not a feature.
//
// foo is the blank scaffold a new domain is copied from; bar next to it is the worked example
// with the bodies filled in. The panics are deliberate — they are what an unimplemented method
// should do in a template, because a silent zero value would let a half-copied domain look like
// it works. Its HTTP routes are not registered; see internal/delivery/http/router/public.go.
//
// The copy procedure is in .claude/skills/crud-operations/SKILL.md.

package foo

type Foo struct {
	ID   string
	Code string
	Foo  string
}

func (e *Foo) Clone() *Foo {
	return &Foo{
		ID:   e.ID,
		Code: e.Code,
		Foo:  e.Foo,
	}
}
