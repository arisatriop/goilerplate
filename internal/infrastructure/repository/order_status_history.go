package repository

import (
	"context"
	"goilerplate/internal/domain/orderstatushistory"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type orderStatusHistoryRepo struct {
	db *gorm.DB
}

func NewOrderStatusHistory(db *gorm.DB) orderstatushistory.Repository {
	return &orderStatusHistoryRepo{db: db}
}

func (r *orderStatusHistoryRepo) WithTx(ctx context.Context) orderstatushistory.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewOrderStatusHistory(tx)
	}
	return r
}

func (r *orderStatusHistoryRepo) CreateOrderStatusHistory(ctx context.Context, history *orderstatushistory.OrderStatusHistory) error {
	osh := &model.OrderStatusHistory{
		ID:        history.ID,
		OrderID:   history.OrderID,
		Status:    history.Status,
		CreatedAt: history.CreatedAt,
	}

	if err := r.db.WithContext(ctx).Create(osh).Error; err != nil {
		return err
	}

	return nil
}

func (r *orderStatusHistoryRepo) toDomainEntity(m *model.OrderStatusHistory) *orderstatushistory.OrderStatusHistory {
	if m == nil {
		return nil
	}

	return &orderstatushistory.OrderStatusHistory{
		ID:        m.ID,
		OrderID:   m.OrderID,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
}
