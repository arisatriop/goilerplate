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

type Usecase interface {
	Create(ctx context.Context, entity *Foo) error
	Update(ctx context.Context, entity *Foo) error
	Delete(ctx context.Context, entity *Foo) error

	GetByID(ctx context.Context, id string) (*Foo, error)
	GetList(ctx context.Context, filter *Filter) ([]*Foo, int64, error)

	BulkCreate(ctx context.Context, entities []*Foo) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Foo) error {
	panic("Implement me")
}

func (uc *usecase) ExistsByCode(ctx context.Context, code string) (bool, error) {
	panic("Implement me")
}

func (uc *usecase) Update(ctx context.Context, entity *Foo) error {
	panic("Implement me")
}

func (uc *usecase) Delete(ctx context.Context, entity *Foo) error {
	panic("Implement me")
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Foo, error) {
	panic("Implement me")
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Foo, int64, error) {
	panic("Implement me")
}

func (uc *usecase) Count(ctx context.Context, filter *Filter) (int64, error) {
	panic("Implement me")
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Foo) error {
	panic("Implement me")
}
