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
