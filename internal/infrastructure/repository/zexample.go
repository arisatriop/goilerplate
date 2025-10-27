package repository

import (
	"context"
	"fmt"

	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type example struct {
	db *gorm.DB
}

func NewExample(db *gorm.DB) zexample.Repository {
	return &example{
		db: db,
	}
}

func (r *example) CreateExample(ctx context.Context, entity *zexample.Example) (*zexample.Example, error) {
	now := utils.TimeJakarta()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model := &model.Example{
		Code:      entity.Code,
		Example:   entity.Example,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: user,
		UpdatedBy: user,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, fmt.Errorf("failed to create example: %w", err)
	}

	return r.modelToEntity(model), nil
}

func (r *example) UpdateExample(ctx context.Context, entity *zexample.Example) error {
	model, err := r.getExampleByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	model.Code = entity.Code
	model.Example = entity.Example
	model.UpdatedAt = utils.TimeJakarta()
	model.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update example: %w", err)
	}

	return nil
}

func (r *example) DeleteExample(ctx context.Context, entity *zexample.Example) error {
	model, err := r.getExampleByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	now := utils.TimeJakarta()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model.DeletedAt = &now
	model.DeletedBy = &user

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to delete example: %w", err)
	}

	return nil
}

func (r *example) GetExampleByID(ctx context.Context, id string) (*zexample.Example, error) {

	model, err := r.getExampleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get example by ID: %w", err)
	}

	return r.modelToEntity(model), nil
}

func (r *example) GetExampleList(ctx context.Context, filter *zexample.Filter) ([]*zexample.Example, error) {
	var models []model.Example

	query := r.db.WithContext(ctx).
		Where("deleted_at IS NULL")

	query = r.applyExampleFilters(query, filter, true) // true = apply pagination

	err := query.Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("")
	}

	entities := make([]*zexample.Example, len(models))
	for i, model := range models {
		entities[i] = r.modelToEntity(&model)
	}

	return entities, nil
}

func (r *example) CountExample(ctx context.Context, filter *zexample.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Example{}).
		Where("deleted_at IS NULL")

	query = r.applyExampleFilters(query, filter, false) // false = don't apply pagination

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count examples: %w", err)
	}

	return count, nil
}

func (r *example) BulkCreate(ctx context.Context, entities []*zexample.Example) error {
	if len(entities) == 0 {
		return nil
	}

	now := utils.TimeJakarta()
	user := ctx.Value(constants.ContextKeyUserID).(string) // Use proper context key

	models := make([]model.Example, len(entities))
	for i, entity := range entities {
		models[i] = model.Example{
			Code:      entity.Code,
			Example:   entity.Example,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: user,
			UpdatedBy: user,
		}
	}

	if err := r.db.WithContext(ctx).Create(&models).
		Select("code, example, is_active, created_at, created_by, updated_at, updated_by").
		Error; err != nil {
		return fmt.Errorf("failed to bulk create examples: %w", err)
	}

	return nil
}

func (r *example) getExampleByID(ctx context.Context, id string) (*model.Example, error) {

	var data model.Example

	err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", id).
		First(&data).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.Error(404, "Example not found")
		}
		return nil, fmt.Errorf("failed to get example by ID: %w", err)
	}

	return &data, nil
}

func (r *example) applyExampleFilters(query *gorm.DB, filter *zexample.Filter, applyPagination bool) *gorm.DB {
	if filter == nil {
		return query
	}

	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("code ILIKE ? OR example ILIKE ?", keyword, keyword)
	}

	if filter.Code != "" {
		query = query.Where("code = ?", filter.Code)
	}

	if applyPagination && filter.Pagination != nil {
		query = query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}

	return query
}

func (r *example) modelToEntity(model *model.Example) *zexample.Example {
	return &zexample.Example{
		ID:      model.ID,
		Code:    model.Code,
		Example: model.Example,
	}
}
