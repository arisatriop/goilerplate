package repository

import (
	"context"
	"goilerplate/internal/domain/plan"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type planRepo struct {
	db *gorm.DB
}

func NewPlan(db *gorm.DB) plan.Repository {
	return &planRepo{
		db: db,
	}
}

func (r *planRepo) WithTx(ctx context.Context) plan.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewPlan(tx)
	}
	return r
}

func (r *planRepo) GetPlanByPlanTypeCode(ctx context.Context, code string) ([]plan.Plan, error) {
	var plans []model.Plan
	if err := r.db.WithContext(ctx).
		Joins("JOIN plan_types ON plan_types.id = plans.plan_type_id").
		Where("plan_types.code = ?", code).
		Where("plan_types.deleted_at IS NULL").
		Where("plans.deleted_at IS NULL").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return r.toDomainEntities(plans), nil
}

func (r *planRepo) toDomainEntities(m []model.Plan) []plan.Plan {
	var plans []plan.Plan
	for i := range m {
		if entity := r.toDomainEntity(&m[i]); entity != nil {
			plans = append(plans, *entity)
		}
	}
	return plans
}

func (r *planRepo) toDomainEntity(m *model.Plan) *plan.Plan {
	if m == nil {
		return nil
	}
	return &plan.Plan{
		ID:             m.ID,
		PlanTypeID:     m.PlanTypeID,
		DurationInDays: m.DurationInDays,
		Price:          m.Price,
		IsActive:       m.IsActive,
	}
}
