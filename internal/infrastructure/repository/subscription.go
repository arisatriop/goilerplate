package repository

import (
	"context"
	"goilerplate/internal/domain/subscription"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscription(db *gorm.DB) subscription.Repository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) WithTx(ctx context.Context) subscription.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewSubscription(tx)
	}
	return r
}

func (r *subscriptionRepo) CreateSubscription(ctx context.Context, sub *subscription.Subscription) (*subscription.Subscription, error) {
	var model model.Subscription
	model.StoreID = sub.StoreID
	model.PlanID = sub.PlanID
	model.StartDate = sub.StarDate
	model.EndDate = sub.EndDate
	model.Price = sub.Price
	model.Status = sub.Status
	model.IsActive = sub.IsActive

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return r.toDomainEntity(&model), nil
}

func (r *subscriptionRepo) toDomainEntity(m *model.Subscription) *subscription.Subscription {
	return &subscription.Subscription{
		ID:       m.ID,
		StoreID:  m.StoreID,
		PlanID:   m.PlanID,
		StarDate: m.StartDate,
		EndDate:  m.EndDate,
		Price:    m.Price,
		Status:   m.Status,
		IsActive: m.IsActive,
	}
}
