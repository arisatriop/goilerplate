package repository

import (
	"context"
	"goilerplate/internal/domain/zexamplenew"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type zexampleNewRepo struct {
	db *gorm.DB
}

func NewZexampleNew(db *gorm.DB) zexamplenew.Repository {
	return &zexampleNewRepo{
		db: db,
	}
}

func (r *zexampleNewRepo) WithTx(ctx context.Context) zexamplenew.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &zexampleNewRepo{db: tx}
	}
	return r
}

func (r *zexampleNewRepo) CreateExample(ctx context.Context, entity *zexamplenew.ZexampleNew) (*zexamplenew.ZexampleNew, error) {
	panic("Implement me")
}

func (r *zexampleNewRepo) UpdateExample(ctx context.Context, entity *zexamplenew.ZexampleNew) error {
	panic("Implement me")
}

func (r *zexampleNewRepo) DeleteExample(ctx context.Context, entity *zexamplenew.ZexampleNew) error {
	panic("Implement me")
}

func (r *zexampleNewRepo) GetExampleByID(ctx context.Context, id string) (*zexamplenew.ZexampleNew, error) {
	panic("Implement me")
}

func (r *zexampleNewRepo) GetExampleList(ctx context.Context, filter *zexamplenew.Filter) ([]*zexamplenew.ZexampleNew, error) {
	panic("Implement me")
}

func (r *zexampleNewRepo) CountExample(ctx context.Context, filter *zexamplenew.Filter) (int64, error) {
	panic("Implement me")
}

func (r *zexampleNewRepo) BulkCreate(ctx context.Context, entities []*zexamplenew.ZexampleNew) error {
	panic("Implement me")
}
