package productcategory

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateProductCategory(ctx context.Context, category *ProductCategory) error

	DeleteProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) error
	DeleteProductCategoryByProductID(ctx context.Context, productID uuid.UUID) error

	GetProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) (*ProductCategory, error)
	GetProductCategoriesByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]ProductCategory, error)
	GetProductCategoriesByProductID(ctx context.Context, productID uuid.UUID) ([]ProductCategory, error)
}
