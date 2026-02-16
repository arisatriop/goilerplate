package product

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	SoftDeleteProduct(ctx context.Context, id uuid.UUID) error
	UpdateToggleActive(ctx context.Context, productID uuid.UUID, isActive bool, updatedBy string) error
	UpdateToggleAvailable(ctx context.Context, productID uuid.UUID, isAvailable bool, updatedBy string) error

	GetProductByID(ctx context.Context, id uuid.UUID, filter *Filter) (*Product, error)
	GetProductList(ctx context.Context, filter *Filter) ([]*Product, error)
	GetMenuProductList(ctx context.Context, filter *Filter) ([]*Product, error)
	GetProductListWithCategories(ctx context.Context, filter *Filter) ([]ProductWithCategories, error)

	CountMenuProductList(ctx context.Context, filter *Filter) (int64, error)
	CountProductListWithCategories(ctx context.Context, filter *Filter) (int64, error)
}
