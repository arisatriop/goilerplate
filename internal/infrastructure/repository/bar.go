package repository

import (
	"context"

	"goilerplate/internal/domain/bar"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

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
		return &barRepo{db: tx}
	}
	return r
}

func (r *barRepo) CreateBar(ctx context.Context, entity *bar.Bar) (*bar.Bar, error) {
	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model := &model.Bar{
		Code:      entity.Code,
		Bar:   entity.Bar,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: user,
		UpdatedBy: user,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, utils.WrapErr(err)
	}

	return r.modelToEntity(model), nil
}

func (r *barRepo) UpdateBar(ctx context.Context, entity *bar.Bar) error {
	model, err := r.getBarByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	model.Code = entity.Code
	model.Bar = entity.Bar
	model.UpdatedAt = utils.Now()
	model.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(model).Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *barRepo) DeleteBar(ctx context.Context, entity *bar.Bar) error {
	model, err := r.getBarByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model.DeletedAt = &now
	model.DeletedBy = &user

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *barRepo) GetBarByID(ctx context.Context, id string) (*bar.Bar, error) {

	model, err := r.getBarByID(ctx, id)
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	return r.modelToEntity(model), nil
}

func (r *barRepo) GetBarList(ctx context.Context, filter *bar.Filter) ([]*bar.Bar, error) {
	var models []model.Bar

	query := r.db.WithContext(ctx).
		Select("id", "code", "bar").
		Where("deleted_at IS NULL")

	r.applyBarFilters(query, filter, true) // true = apply pagination

	err := query.Find(&models).Error
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	entities := make([]*bar.Bar, len(models))
	for i, model := range models {
		entities[i] = r.modelToEntity(&model)
	}

	return entities, nil
}

func (r *barRepo) CountBar(ctx context.Context, filter *bar.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Bar{}).
		Where("deleted_at IS NULL")

	r.applyBarFilters(query, filter, false) // false = don't apply pagination

	if err := query.Count(&count).Error; err != nil {
		return 0, utils.WrapErr(err)
	}

	return count, nil
}

func (r *barRepo) BulkCreate(ctx context.Context, entities []*bar.Bar) error {
	if len(entities) == 0 {
		return nil
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string) // Use proper context key

	models := make([]model.Bar, len(entities))
	for i, entity := range entities {
		models[i] = model.Bar{
			Code:      entity.Code,
			Bar:   entity.Bar,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: user,
			UpdatedBy: user,
		}
	}

	if err := r.db.WithContext(ctx).Create(&models).
		Select("code, bar, is_active, created_at, created_by, updated_at, updated_by").
		Error; err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func (r *barRepo) getBarByID(ctx context.Context, id string) (*model.Bar, error) {

	var data model.Bar

	err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", id).
		First(&data).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ClientErr(404, "Bar not found")
		}
		return nil, utils.WrapErr(err)
	}

	return &data, nil
}

func (r *barRepo) applyBarFilters(query *gorm.DB, filter *bar.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query.Where("code ILIKE ? OR bar ILIKE ?", keyword, keyword)
	}

	if filter.Code != "" {
		query.Where("code = ?", filter.Code)
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *barRepo) modelToEntity(model *model.Bar) *bar.Bar {
	return &bar.Bar{
		ID:      model.ID,
		Code:    model.Code,
		Bar: model.Bar,
	}
}
