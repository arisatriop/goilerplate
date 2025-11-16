package repository

import (
	"context"
	"goilerplate/internal/domain/bar"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type barRepo struct {
	db *gorm.DB
}

func NewBar(db *gorm.DB) bar.Repository {
	return &barRepo{
		db: db,
	}
}

func (r *barRepo) WithTx(ctx context.Context) bar.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewBar(tx)
	}
	return r
}

// Implement other methods like CreateBar, UpdateBar, DeleteBar, SoftDeleteBar, GetBarList, GetBarByID, CountBars here.
