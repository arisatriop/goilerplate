package repository

import (
	"context"
	"goilerplate/internal/domain/category"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type categoryRepo struct {
	db *gorm.DB
}

func NewCategory(db *gorm.DB) category.Repository {
	return &categoryRepo{
		db: db,
	}
}

func (r *categoryRepo) WithTx(ctx context.Context) category.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewCategory(tx)
	}
	return r
}

func (r *categoryRepo) GetCategoryListWithProducts(ctx context.Context, filter *category.CategoryWithProductsFilter) ([]category.CategoryWithProducts, error) {

	var results []category.CategoryWithProducts
	query := r.db.WithContext(ctx).
		Table("categories").
		Select(`
			categories.id as id,
			categories.name as name,
			products.id as product_id,
			products.name as product_name,
			products.description as product_desc,
			products.price as product_price,
			products.images as product_images,
			products.is_available as product_is_available
		`).
		Joins("LEFT JOIN product_categories ON categories.id = product_categories.category_id").
		Joins("LEFT JOIN products ON product_categories.product_id = products.id AND products.is_active = 1 AND products.deleted_at IS NULL").
		Where("categories.store_id = ? AND categories.is_active = 1 AND categories.deleted_at IS NULL", filter.StoreID)

	if filter.CategoryID != nil {
		query = query.Where("categories.id = ?", *filter.CategoryID)
	}

	if filter.Keyword != "" {
		query = query.Where("LOWER(products.name) LIKE LOWER(?)", "%"+filter.Keyword+"%")
	}

	query = query.Order("categories.created_at, products.name")

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (r *categoryRepo) CreateCategory(ctx context.Context, category *category.Category) error {
	model := &model.Category{
		ID:       category.ID,
		Name:     category.Name,
		StoreID:  category.StoreID,
		IsActive: category.IsActive,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *categoryRepo) SoftDeleteCategory(ctx context.Context, id uuid.UUID, deletedBy string) error {
	updates := map[string]interface{}{
		"deleted_at": utils.Now(),
		"deleted_by": deletedBy,
	}

	return r.db.WithContext(ctx).
		Model(&model.Category{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *categoryRepo) GetCategoryList(ctx context.Context, filter *category.Filter) ([]*category.Category, error) {
	var models []model.Category

	query := r.db.WithContext(ctx).
		Select("id", "name", "store_id", "is_active").
		Where("deleted_at IS NULL").
		Order("name DESC")

	r.applyCategoryFilters(query, filter, true)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *categoryRepo) GetCategoryByID(ctx context.Context, id uuid.UUID, filter *category.Filter) (*category.Category, error) {
	var m model.Category

	query := r.db.WithContext(ctx).
		Select("id", "name", "store_id", "is_active").
		Where("id = ? AND deleted_at IS NULL", id)

	r.applyCategoryFilters(query, filter, false)

	if err := query.First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntity(&m), nil
}

func (r *categoryRepo) UpdateToggleActive(ctx context.Context, category *category.Category, updatedBy string) error {
	updates := map[string]interface{}{
		"is_active":  category.IsActive,
		"updated_by": updatedBy,
	}

	return r.db.WithContext(ctx).
		Model(&model.Category{}).
		Where("id = ?", category.ID).
		Updates(updates).Error
}

func (r *categoryRepo) CountCategories(ctx context.Context, filter *category.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Category{}).
		Where("deleted_at IS NULL")

	r.applyCategoryFilters(query, filter, false)

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *categoryRepo) GetCategoryByProductID(ctx context.Context, id uuid.UUID) ([]category.Category, error) {
	var m []model.Category

	query := r.db.WithContext(ctx).
		Select("categories.*").
		Table("categories").
		Joins("JOIN product_categories ON product_categories.category_id = categories.id").
		Joins("JOIN products ON products.id = product_categories.product_id").
		Where("products.id = ? AND categories.deleted_at IS NULL", id)

	if err := query.Find(&m).Error; err != nil {
		return nil, err
	}

	res := r.toDomainEntities(m)
	result := make([]category.Category, len(res))
	for i, v := range res {
		result[i] = *v
	}

	return result, nil

}

func (r *categoryRepo) applyCategoryFilters(query *gorm.DB, filter *category.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.StoreID != nil {
		query.Where("store_id = ?", *filter.StoreID)
	}

	if filter.IsActive != nil {
		query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Keyword != "" {
		query.Where("LOWER(name) LIKE LOWER(?)", "%"+filter.Keyword+"%")
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *categoryRepo) toDomainEntities(m []model.Category) []*category.Category {
	var categories []*category.Category
	for i := range m {
		if entity := r.toDomainEntity(&m[i]); entity != nil {
			categories = append(categories, entity)
		}
	}

	return categories
}

func (r *categoryRepo) toDomainEntity(m *model.Category) *category.Category {
	if m == nil {
		return nil
	}

	return &category.Category{
		ID:       m.ID,
		Name:     m.Name,
		StoreID:  m.StoreID,
		IsActive: m.IsActive,
	}
}
