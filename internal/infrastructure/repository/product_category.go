package repository

import (
	"context"
	"goilerplate/internal/domain/productcategory"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productCategoryRepo struct {
	db *gorm.DB
}

func NewProductCategory(db *gorm.DB) productcategory.Repository {
	return &productCategoryRepo{db: db}
}

func (r *productCategoryRepo) WithTx(ctx context.Context) productcategory.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewProductCategory(tx)
	}
	return r
}

func (r *productCategoryRepo) DeleteProductCategoryByProductID(ctx context.Context, productID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Delete(&model.ProductCategory{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productCategoryRepo) DeleteProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("product_id = ? AND category_id = ?", productID, categoryID).
		Delete(&model.ProductCategory{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productCategoryRepo) GetProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) (*productcategory.ProductCategory, error) {
	var m model.ProductCategory
	if err := r.db.WithContext(ctx).
		Where("product_id = ? AND category_id = ?", productID, categoryID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	entity := r.toDomainEntity(&m)
	return &entity, nil
}

func (r *productCategoryRepo) CreateProductCategory(ctx context.Context, pc *productcategory.ProductCategory) error {
	model := &model.ProductCategory{
		ProductID:  pc.ProductID,
		CategoryID: pc.CategoryID,
		IsActive:   pc.IsActive,
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(model).Error; err != nil {
		return err
	}

	return nil
}

func (r *productCategoryRepo) GetProductCategoriesByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]productcategory.ProductCategory, error) {
	var models []model.ProductCategory
	if err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Find(&models).Error; err != nil {

		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *productCategoryRepo) GetProductCategoriesByProductID(ctx context.Context, productID uuid.UUID) ([]productcategory.ProductCategory, error) {
	var models []model.ProductCategory
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Find(&models).Error; err != nil {

		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *productCategoryRepo) toDomainEntities(models []model.ProductCategory) []productcategory.ProductCategory {
	entities := make([]productcategory.ProductCategory, len(models))
	for i, m := range models {
		entities[i] = r.toDomainEntity(&m)
	}
	return entities
}

func (r *productCategoryRepo) toDomainEntity(m *model.ProductCategory) productcategory.ProductCategory {
	return productcategory.ProductCategory{
		ProductID:  m.ProductID,
		CategoryID: m.CategoryID,
		IsActive:   m.IsActive,
	}
}
