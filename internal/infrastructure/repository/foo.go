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
