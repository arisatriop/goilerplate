package repository

import (
	"context"

	"goilerplate/internal/domain/bazs"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type bazsRepo struct {
	db *gorm.DB
}

func NewBazs(db *gorm.DB) bazs.Repository {
	return &bazsRepo{
		db: db,
	}
}

func (r *bazsRepo) WithTx(ctx context.Context) bazs.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &bazsRepo{db: tx}
	}
	return r
}

func (r *bazsRepo) CreateBazs(ctx context.Context, entity *bazs.Bazs) (*bazs.Bazs, error) {
	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	m := &model.Bazs{
		Code:      entity.Code,
		Name:      entity.Name,
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

func (r *bazsRepo) UpdateBazs(ctx context.Context, entity *bazs.Bazs) error {
	m, err := r.getBazsByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	m.Code = entity.Code
	m.Name = entity.Name
	m.UpdatedAt = utils.Now()
	m.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(m).Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *bazsRepo) DeleteBazs(ctx context.Context, entity *bazs.Bazs) error {
	m, err := r.getBazsByID(ctx, entity.ID)
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

func (r *bazsRepo) GetBazsByID(ctx context.Context, id string) (*bazs.Bazs, error) {
	m, err := r.getBazsByID(ctx, id)
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	return r.modelToEntity(m), nil
}

func (r *bazsRepo) GetBazsList(ctx context.Context, filter *bazs.Filter) ([]*bazs.Bazs, error) {
	var models []model.Bazs

	query := r.db.WithContext(ctx).
		Select("id", "code", "name").
		Where("deleted_at IS NULL")

	r.applyBazsFilters(query, filter, true)

	err := query.Find(&models).Error
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	entities := make([]*bazs.Bazs, len(models))
	for i, m := range models {
		entities[i] = r.modelToEntity(&m)
	}

	return entities, nil
}

func (r *bazsRepo) CountBazs(ctx context.Context, filter *bazs.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Bazs{}).
		Where("deleted_at IS NULL")

	r.applyBazsFilters(query, filter, false)

	if err := query.Count(&count).Error; err != nil {
		return 0, utils.WrapErr(err)
	}

	return count, nil
}

func (r *bazsRepo) BulkCreate(ctx context.Context, entities []*bazs.Bazs) error {
	if len(entities) == 0 {
		return nil
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)

	models := make([]model.Bazs, len(entities))
	for i, entity := range entities {
		models[i] = model.Bazs{
			Code:      entity.Code,
			Name:      entity.Name,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: user,
			UpdatedBy: user,
		}
	}

	if err := r.db.WithContext(ctx).Create(&models).
		Select("code, name, is_active, created_at, created_by, updated_at, updated_by").
		Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *bazsRepo) getBazsByID(ctx context.Context, id string) (*model.Bazs, error) {
	var data model.Bazs

	err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", id).
		First(&data).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ClientErr(404, "Bazs not found")
		}
		return nil, utils.WrapErr(err)
	}

	return &data, nil
}

func (r *bazsRepo) applyBazsFilters(query *gorm.DB, filter *bazs.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query.Where("code ILIKE ? OR name ILIKE ?", keyword, keyword)
	}

	if filter.Code != "" {
		query.Where("code = ?", filter.Code)
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *bazsRepo) modelToEntity(m *model.Bazs) *bazs.Bazs {
	return &bazs.Bazs{
		ID:   m.ID,
		Code: m.Code,
		Name: m.Name,
	}
}
