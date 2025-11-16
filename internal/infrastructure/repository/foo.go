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
		return NewFoo(tx)
	}
	return r
}

// Implement other methods like CreateFoo, UpdateFoo, DeleteFoo, SoftDeleteFoo, GetFooList, GetFooByID, CountFoos here.
