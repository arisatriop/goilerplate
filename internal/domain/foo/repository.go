// This file is a TEMPLATE, not a feature.
//
// foo is the blank scaffold a new domain is copied from; bar next to it is the worked example
// with the bodies filled in. The panics are deliberate — they are what an unimplemented method
// should do in a template, because a silent zero value would let a half-copied domain look like
// it works. Its HTTP routes are not registered; see internal/delivery/http/router/public.go.
//
// The copy procedure is in .claude/skills/crud-operations/SKILL.md.

package foo

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateFoo(ctx context.Context, entities *Foo) (*Foo, error)
	UpdateFoo(ctx context.Context, entities *Foo) error
	DeleteFoo(ctx context.Context, entities *Foo) error
	BulkCreate(ctx context.Context, entities []*Foo) error

	CountFoo(ctx context.Context, filter *Filter) (int64, error)
	GetFooList(ctx context.Context, filter *Filter) ([]*Foo, error)
	GetFooByID(ctx context.Context, id string) (*Foo, error)
}
