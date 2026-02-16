package repository

import (
	"context"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type planTypeRepo struct {
	db *gorm.DB
}

func NewPlanType(db *gorm.DB) plantype.Repository {
	return &planTypeRepo{db: db}
}

func (r *planTypeRepo) WithTx(ctx context.Context) plantype.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewPlanType(tx)
	}
	return r
}

func (r *planTypeRepo) GetStoreActiveSubsciption(ctx context.Context, storeID uuid.UUID) ([]plantype.PlanType, error) {
	var pt []model.PlanType
	if err := r.db.WithContext(ctx).
		Raw(
			`SELECT
				pt.id,
				pt.code,
				pt.name,
				pt.is_active,
				pt.created_at,
				pt.updated_at,
				pt.deleted_at,
				pt.created_by,
				pt.updated_by,
				pt.deleted_by
			FROM plan_types pt
			JOIN plans p ON p.plan_type_id = pt.id
			JOIN subscriptions s ON s.plan_id = p.id
			WHERE s.store_id = ? AND s.is_active = TRUE
			AND s.deleted_at IS NULL
			AND (s.end_date > NOW() OR s.end_date IS NULL)`, storeID,
		).
		Scan(&pt).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntities(pt), nil
}

func (r *planTypeRepo) toDomainEntities(m []model.PlanType) []plantype.PlanType {
	var planTypes []plantype.PlanType
	for i := range m {
		if entity := r.toDomainEntity(&m[i]); entity != nil {
			planTypes = append(planTypes, *entity)
		}
	}
	return planTypes
}

func (r *planTypeRepo) toDomainEntity(m *model.PlanType) *plantype.PlanType {
	if m == nil {
		return nil
	}
	return &plantype.PlanType{
		ID:       m.ID,
		Code:     m.Code,
		Name:     m.Name,
		IsActive: m.IsActive,
	}
}

type planTypeWithPlanRow struct {
	PlanTypeID       uuid.UUID
	PlanTypeCode     string
	PlanTypeName     string
	PlanTypeIsActive bool
	PlanID           *uuid.UUID
	DurationInDays   *int
	Price            *string
	PlanIsActive     *bool
}

func (r *planTypeRepo) GetListWithPlans(ctx context.Context) ([]plantype.PlanTypeWithPlans, error) {

	var rows []planTypeWithPlanRow
	err := r.db.WithContext(ctx).
		Table("plan_types pt").
		Select(`
			pt.id as plan_type_id,
			pt.code as plan_type_code,
			pt.name as plan_type_name,
			pt.is_active as plan_type_is_active,
			p.id as plan_id,
			p.duration_in_days,
			p.price,
			p.is_active as plan_is_active
		`).
		Joins("LEFT JOIN plans p ON p.plan_type_id = pt.id AND p.deleted_at IS NULL").
		Where("pt.deleted_at IS NULL").
		Where("pt.is_active = ?", true).
		Order("pt.code, p.duration_in_days").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	return r.groupPlanTypeWithPlans(rows), nil
}

func (r *planTypeRepo) groupPlanTypeWithPlans(rows []planTypeWithPlanRow) []plantype.PlanTypeWithPlans {
	planTypeMap := make(map[uuid.UUID]*plantype.PlanTypeWithPlans)
	var planTypeOrder []uuid.UUID

	for _, row := range rows {
		if _, exists := planTypeMap[row.PlanTypeID]; !exists {
			planTypeMap[row.PlanTypeID] = &plantype.PlanTypeWithPlans{
				ID:       row.PlanTypeID,
				Code:     row.PlanTypeCode,
				Name:     row.PlanTypeName,
				IsActive: row.PlanTypeIsActive,
				Plans:    []plantype.PlanItem{},
			}
			planTypeOrder = append(planTypeOrder, row.PlanTypeID)
		}

		if row.PlanID != nil && row.DurationInDays != nil && row.Price != nil && row.PlanIsActive != nil {
			price, _ := decimal.NewFromString(*row.Price)
			planTypeMap[row.PlanTypeID].Plans = append(planTypeMap[row.PlanTypeID].Plans, plantype.PlanItem{
				ID:             *row.PlanID,
				DurationInDays: *row.DurationInDays,
				Price:          price,
				IsActive:       *row.PlanIsActive,
			})
		}
	}

	result := make([]plantype.PlanTypeWithPlans, 0, len(planTypeMap))
	for _, id := range planTypeOrder {
		result = append(result, *planTypeMap[id])
	}

	return result
}
