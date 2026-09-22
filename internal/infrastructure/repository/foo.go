// This file is a TEMPLATE, not a feature.
//
// foo is the blank scaffold a new domain is copied from; bar next to it is the worked example
// with the bodies filled in. The panics are deliberate — they are what an unimplemented method
// should do in a template, because a silent zero value would let a half-copied domain look like
// it works. Its HTTP routes are not registered; see internal/delivery/http/router/public.go.
//
// The copy procedure is in .claude/skills/crud-operations/SKILL.md.

package repository

import (
	"context"
	"goilerplate/internal/domain/foo"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type fooRepo struct {
	db *gorm.DB
}

func NewFoo(db *gorm.DB) foo.Repository {
	return &fooRepo{
		db: db,
	}
}

func (r *fooRepo) WithTx(ctx context.Context) foo.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &fooRepo{db: tx}
	}
	return r
}

func (r *fooRepo) CreateFoo(ctx context.Context, entity *foo.Foo) (*foo.Foo, error) {
	panic("Implement me")
}

func (r *fooRepo) UpdateFoo(ctx context.Context, entity *foo.Foo) error {
	panic("Implement me")
}

func (r *fooRepo) DeleteFoo(ctx context.Context, entity *foo.Foo) error {
	panic("Implement me")
}

func (r *fooRepo) GetFooByID(ctx context.Context, id string) (*foo.Foo, error) {
	panic("Implement me")
}

func (r *fooRepo) GetFooList(ctx context.Context, filter *foo.Filter) ([]*foo.Foo, error) {
	panic("Implement me")
}

func (r *fooRepo) CountFoo(ctx context.Context, filter *foo.Filter) (int64, error) {
	panic("Implement me")
}

func (r *fooRepo) BulkCreate(ctx context.Context, entities []*foo.Foo) error {
	panic("Implement me")
}
