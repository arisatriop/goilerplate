package repository

import (
	"context"

	"goilerplate/internal/domain/bas"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type basRepo struct {
	db *gorm.DB
}

func NewBas(db *gorm.DB) bas.Repository {
	return &basRepo{db: db}
}

func (r *basRepo) WithTx(ctx context.Context) bas.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &basRepo{db: tx}
	}
	return r
}

func (r *basRepo) CreateBas(ctx context.Context, entity *bas.Bas) (*bas.Bas, error) {
	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	m := &model.Bas{
		Code:      entity.Code,
		Bas:       entity.Bas,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: user,
		UpdatedBy: user,
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, utils.WrapErr(err)
	}

	return r.modelToEntity(m), nil
}

func (r *basRepo) UpdateBas(ctx context.Context, entity *bas.Bas) error {
	m, err := r.getBasByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	m.Code = entity.Code
	m.Bas = entity.Bas
	m.UpdatedAt = utils.Now()
	m.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(m).Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *basRepo) DeleteBas(ctx context.Context, entity *bas.Bas) error {
	m, err := r.getBasByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	m.DeletedAt = &now
	m.DeletedBy = &user

	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *basRepo) GetBasByID(ctx context.Context, id string) (*bas.Bas, error) {
	m, err := r.getBasByID(ctx, id)
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	return r.modelToEntity(m), nil
}

func (r *basRepo) GetBasList(ctx context.Context, filter *bas.Filter) ([]*bas.Bas, error) {
	var models []model.Bas

	query := r.db.WithContext(ctx).
		Select("id", "code", "bas").
		Where("deleted_at IS NULL")

	r.applyBasFilters(query, filter, true)

	if err := query.Find(&models).Error; err != nil {
		return nil, utils.WrapErr(err)
	}

	entities := make([]*bas.Bas, len(models))
	for i, m := range models {
		entities[i] = r.modelToEntity(&m)
	}

	return entities, nil
}

func (r *basRepo) CountBas(ctx context.Context, filter *bas.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Bas{}).
		Where("deleted_at IS NULL")

	r.applyBasFilters(query, filter, false)

	if err := query.Count(&count).Error; err != nil {
		return 0, utils.WrapErr(err)
	}

	return count, nil
}

func (r *basRepo) getBasByID(ctx context.Context, id string) (*model.Bas, error) {
	var data model.Bas

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&data).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ClientErr(404, "Bas not found")
		}
		return nil, utils.WrapErr(err)
	}

	return &data, nil
}

func (r *basRepo) applyBasFilters(query *gorm.DB, filter *bas.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query.Where("code ILIKE ? OR bas ILIKE ?", keyword, keyword)
	}

	if filter.Code != "" {
		query.Where("code = ?", filter.Code)
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *basRepo) modelToEntity(m *model.Bas) *bas.Bas {
	return &bas.Bas{
		ID:   m.ID,
		Code: m.Code,
		Bas:  m.Bas,
	}
}
