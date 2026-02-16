package repository

import (
	"context"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productRepo struct {
	db *gorm.DB
}

func NewProduct(db *gorm.DB) product.Repository {
	return &productRepo{db: db}
}

func (r *productRepo) WithTx(ctx context.Context) product.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewProduct(tx)
	}
	return r
}

// * COMMAND
func (r *productRepo) CreateProduct(ctx context.Context, prod *product.Product) error {
	p := &model.Product{
		ID:          prod.ID,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Images:      prod.Images,
		IsActive:    prod.IsActive,
		StoreID:     prod.StoreID,
		IsAvailable: prod.IsAvailable,
	}

	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, prod *product.Product) error {
	// Fetch the existing product first
	var existingProduct model.Product
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", prod.ID).First(&existingProduct).Error; err != nil {
		return err
	}

	// Update only name, description, and price
	existingProduct.Name = prod.Name
	existingProduct.Description = prod.Description
	existingProduct.Price = prod.Price
	existingProduct.Images = prod.Images

	// Save (this triggers BeforeUpdate hook)
	if err := r.db.WithContext(ctx).Save(&existingProduct).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepo) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Product{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productRepo) SoftDeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": utils.Now(),
		"deleted_by": ctx.Value(constants.ContextKeyUserID).(string),
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productRepo) UpdateToggleActive(ctx context.Context, productID uuid.UUID, isActive bool, updatedBy string) error {
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
		"is_active":  isActive,
		"updated_by": updatedBy,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productRepo) UpdateToggleAvailable(ctx context.Context, productID uuid.UUID, isAvailable bool, updatedBy string) error {
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
		"is_available": isAvailable,
		"updated_by":   updatedBy,
	}).Error; err != nil {
		return err
	}
	return nil
}

// * END OF COMMAND

// ===

// * QUERY
func (r *productRepo) GetProductByID(ctx context.Context, id uuid.UUID, filter *product.Filter) (*product.Product, error) {
	var model model.Product

	query := r.db.WithContext(ctx).
		Select("id", "name", "description", "price", "images", "store_id", "is_active", "is_available").
		Where("id = ? AND deleted_at IS NULL", id)

	r.applyProductFilters(query, filter, false)

	if err := query.First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntity(&model), nil
}

func (r *productRepo) GetProductList(ctx context.Context, filter *product.Filter) ([]*product.Product, error) {
	var models []*model.Product

	query := r.db.WithContext(ctx).
		Select("id", "name", "description", "price", "images", "store_id", "is_active", "is_available").
		Where("deleted_at IS NULL")

	r.applyProductFilters(query, filter, true)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *productRepo) CountProductList(ctx context.Context, filter *product.Filter) (int64, error) {
	return 0, nil
}

func (r *productRepo) GetMenuProductList(ctx context.Context, filter *product.Filter) ([]*product.Product, error) {
	var models []*model.Product

	query := r.db.WithContext(ctx).
		Select("products.id", "products.name", "products.description", "products.price", "products.images", "products.store_id", "products.is_active", "products.is_available").
		Where("store_id = ? AND products.is_active = ? AND deleted_at IS NULL", filter.StoreID, true)

	if filter.CategoryID != nil {
		query.Joins("JOIN product_categories pc ON products.id = pc.product_id").
			Where("pc.category_id = ?", *filter.CategoryID)
	}

	if filter.Keyword != "" {
		likePattern := "%" + strings.ToLower(filter.Keyword) + "%"
		query.Where("LOWER(products.name) LIKE ?", likePattern)
	}

	if filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *productRepo) CountMenuProductList(ctx context.Context, filter *product.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.Product{}).
		Where("store_id = ? AND products.is_active = ? AND deleted_at IS NULL", filter.StoreID, true)

	if filter.CategoryID != nil {
		query.Joins("JOIN product_categories pc ON products.id = pc.product_id").
			Where("pc.category_id = ?", *filter.CategoryID)
	}

	if filter.Keyword != "" {
		likePattern := "%" + strings.ToLower(filter.Keyword) + "%"
		query.Where("LOWER(products.name) LIKE ?", likePattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *productRepo) GetProductListWithCategories(ctx context.Context, filter *product.Filter) ([]product.ProductWithCategories, error) {
	// Step 1: Get product IDs with proper pagination (no joins to avoid duplication)
	var productIDs []uuid.UUID

	productQuery := r.db.WithContext(ctx).
		Model(&model.Product{}).
		Select("id").
		Where("deleted_at IS NULL")

	if filter.StoreID != nil {
		productQuery.Where("store_id = ?", *filter.StoreID)
	}

	if filter.IsActive != nil {
		productQuery.Where("is_active = ?", *filter.IsActive)
	}

	// For category filtering, use EXISTS to filter products
	if filter.CategoryID != nil {
		productQuery.Where("EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = products.id AND pc.category_id = ?)", *filter.CategoryID)
	}

	if filter.Keyword != "" {
		productQuery.Where(
			"(LOWER(name) LIKE LOWER(?) OR EXISTS (SELECT 1 FROM product_categories pc JOIN categories c ON pc.category_id = c.id WHERE pc.product_id = products.id AND LOWER(c.name) LIKE LOWER(?) AND c.deleted_at IS NULL))",
			"%"+filter.Keyword+"%",
			"%"+filter.Keyword+"%",
		)
	}

	// Apply pagination to products, not joined rows
	if filter.Pagination != nil {
		productQuery.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}

	if err := productQuery.Find(&productIDs).Error; err != nil {
		return nil, err
	}

	if len(productIDs) == 0 {
		return []product.ProductWithCategories{}, nil
	}

	// Step 2: Get full product and category data for the paginated product IDs
	type ProductCategoryResult struct {
		ProductID          uuid.UUID
		ProductName        string
		ProductDescription *string
		ProductPrice       string
		ProductImages      *string
		ProductIsAvailable bool
		ProductIsActive    bool
		CategoryID         *uuid.UUID
		CategoryName       *string
		CategoryIsActive   *bool
		CategoryStoreID    *uuid.UUID
	}

	var results []ProductCategoryResult

	query := r.db.WithContext(ctx).
		Table("products").
		Select(`
			products.id as product_id,
			products.name as product_name,
			products.description as product_description,
			products.price as product_price,
			products.images as product_images,
			products.is_available as product_is_available,
			products.is_active as product_is_active,
			categories.id as category_id,
			categories.name as category_name,
			categories.is_active as category_is_active,
			categories.store_id as category_store_id
		`).
		Joins("LEFT JOIN product_categories ON products.id = product_categories.product_id").
		Joins("LEFT JOIN categories ON product_categories.category_id = categories.id AND categories.deleted_at IS NULL").
		Where("products.id IN ?", productIDs).
		Order("products.id, categories.name")

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	// Step 3: Group results by product
	productMap := make(map[uuid.UUID]*product.ProductWithCategories)
	for _, result := range results {
		if _, exists := productMap[result.ProductID]; !exists {
			productMap[result.ProductID] = &product.ProductWithCategories{
				Product: product.Product{
					ID:          result.ProductID,
					Name:        result.ProductName,
					Description: result.ProductDescription,
					Price:       utils.ParseDecimal(result.ProductPrice),
					Images:      result.ProductImages,
					IsAvailable: result.ProductIsAvailable,
					IsActive:    result.ProductIsActive,
				},
				Categories: []product.Category{},
			}
		}

		// Add category if it exists (not null from LEFT JOIN)
		if result.CategoryID != nil {
			productMap[result.ProductID].Categories = append(
				productMap[result.ProductID].Categories,
				product.Category{
					ID:   *result.CategoryID,
					Name: *result.CategoryName,
				},
			)
		}
	}

	// Step 4: Convert map to slice maintaining the original order from productIDs
	productList := make([]product.ProductWithCategories, 0, len(productIDs))
	for _, productID := range productIDs {
		if prod, exists := productMap[productID]; exists {
			productList = append(productList, *prod)
		}
	}

	return productList, nil
}

func (r *productRepo) CountProductListWithCategories(ctx context.Context, filter *product.Filter) (int64, error) {
	var count int64

	// Use a simpler approach: count products with EXISTS for category filtering
	query := r.db.WithContext(ctx).
		Model(&model.Product{}).
		Where("deleted_at IS NULL")

	if filter.StoreID != nil {
		query.Where("store_id = ?", *filter.StoreID)
	}

	if filter.IsActive != nil {
		query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.CategoryID != nil {
		query.Where("EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = products.id AND pc.category_id = ?)", *filter.CategoryID)
	}

	if filter.Keyword != "" {
		query.Where(
			"(LOWER(name) LIKE LOWER(?) OR EXISTS (SELECT 1 FROM product_categories pc JOIN categories c ON pc.category_id = c.id WHERE pc.product_id = products.id AND LOWER(c.name) LIKE LOWER(?) AND c.deleted_at IS NULL))",
			"%"+filter.Keyword+"%",
			"%"+filter.Keyword+"%",
		)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *productRepo) applyProductFilters(query *gorm.DB, filter *product.Filter, applyPagination bool) {
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
		query.Where("lower(name) = ?", strings.ToLower(filter.Keyword))
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

// * END OF QUERY

// ===

// * CONVERTER

func (r *productRepo) toDomainEntities(models []*model.Product) []*product.Product {
	entities := make([]*product.Product, len(models))
	for i, m := range models {
		entities[i] = r.toDomainEntity(m)
	}
	return entities
}

func (r *productRepo) toDomainEntity(m *model.Product) *product.Product {
	return &product.Product{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Images:      m.Images,
		Price:       m.Price,
		StoreID:     m.StoreID,
		IsActive:    m.IsActive,
		IsAvailable: m.IsAvailable,
	}
}

// * END OF CONVERTER
