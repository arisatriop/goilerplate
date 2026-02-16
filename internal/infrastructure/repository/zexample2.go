package repository

import (
	"context"
	"fmt"

	"goilerplate/internal/domain/zexample2"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type example2 struct {
	db *gorm.DB
}

func NewExample2(db *gorm.DB) zexample2.Repository {
	return &example2{
		db: db,
	}
}

func (r *example2) WithTx(ctx context.Context) zexample2.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &example2{db: tx}
	}
	return r
}

func (r *example2) CreateExample(ctx context.Context, entity *zexample2.Example2) (*zexample2.Example2, error) {
	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model := &model.Example2{
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

func (r *example2) UpdateExample(ctx context.Context, entity *zexample2.Example2) error {
	model, err := r.getExampleByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	model.Code = entity.Code
	model.Example = entity.Example
	model.UpdatedAt = utils.Now()
	model.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update example: %w", err)
	}

	return nil
}

func (r *example2) DeleteExample(ctx context.Context, entity *zexample2.Example2) error {
	model, err := r.getExampleByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model.DeletedAt = &now
	model.DeletedBy = &user

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to delete example: %w", err)
	}

	return nil
}

func (r *example2) GetExampleByID(ctx context.Context, id string) (*zexample2.Example2, error) {

	model, err := r.getExampleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get example by ID: %w", err)
	}

	return r.modelToEntity(model), nil
}

func (r *example2) GetExampleList(ctx context.Context, filter *zexample2.Filter) ([]*zexample2.Example2, error) {
	var models []model.Example2

	query := r.db.WithContext(ctx).
		Select("id", "code", "example").
		Where("deleted_at IS NULL")

	query = r.applyExampleFilters(query, filter, true) // true = apply pagination

	err := query.Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("")
	}

	entities := make([]*zexample2.Example2, len(models))
	for i, model := range models {
		entities[i] = r.modelToEntity(&model)
	}

	return entities, nil
}

func (r *example2) CountExample(ctx context.Context, filter *zexample2.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Example2{}).
		Where("deleted_at IS NULL")

	query = r.applyExampleFilters(query, filter, false) // false = don't apply pagination

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count examples: %w", err)
	}

	return count, nil
}

func (r *example2) BulkCreate(ctx context.Context, entities []*zexample2.Example2) error {
	if len(entities) == 0 {
		return nil
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string) // Use proper context key

	models := make([]model.Example2, len(entities))
	for i, entity := range entities {
		models[i] = model.Example2{
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

func (r *example2) getExampleByID(ctx context.Context, id string) (*model.Example2, error) {

	var data model.Example2

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

func (r *example2) applyExampleFilters(query *gorm.DB, filter *zexample2.Filter, applyPagination bool) *gorm.DB {
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

func (r *example2) modelToEntity(model *model.Example2) *zexample2.Example2 {
	return &zexample2.Example2{
		ID:      model.ID,
		Code:    model.Code,
		Example: model.Example,
	}
}
