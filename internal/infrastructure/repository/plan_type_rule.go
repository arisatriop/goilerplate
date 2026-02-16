package repository

import (
	"context"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type planTypeRuleRepository struct {
	db *gorm.DB
}

func NewPlanTypeRuleRepository(db *gorm.DB) plantyperule.Repository {
	return &planTypeRuleRepository{db: db}
}

func (r *planTypeRuleRepository) WithTx(ctx context.Context) plantyperule.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewPlanTypeRuleRepository(tx)
	}
	return r
}

func (r *planTypeRuleRepository) GetPlanTypeRuleByPlanTypeID(ctx context.Context, planTypeID string) ([]plantyperule.PlanTypeRule, error) {
	var models []model.PlanTypeRule
	if err := r.db.WithContext(ctx).
		Where("plan_type_id = ?", planTypeID).
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Find(&models).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *planTypeRuleRepository) toDomainEntities(models []model.PlanTypeRule) []plantyperule.PlanTypeRule {
	entities := make([]plantyperule.PlanTypeRule, len(models))
	for i, m := range models {
		entities[i] = r.toDomainEntity(&m)
	}
	return entities
}

func (r *planTypeRuleRepository) toDomainEntity(m *model.PlanTypeRule) plantyperule.PlanTypeRule {
	return plantyperule.PlanTypeRule{
		ID:         m.ID,
		PlanTypeID: m.PlanTypeID,
		Rule:       m.Rule,
		RuleValue:  m.RuleValue,
		IsActive:   m.IsActive,
	}
}
