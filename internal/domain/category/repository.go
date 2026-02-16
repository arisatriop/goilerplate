package category

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateCategory(ctx context.Context, category *Category) error
	SoftDeleteCategory(ctx context.Context, id uuid.UUID, deletedBy string) error
	UpdateToggleActive(ctx context.Context, category *Category, updatedBy string) error

	GetCategoryByID(ctx context.Context, id uuid.UUID, filter *Filter) (*Category, error)
	GetCategoryByProductID(ctx context.Context, id uuid.UUID) ([]Category, error)
	CountCategories(ctx context.Context, filter *Filter) (int64, error)

	GetCategoryList(ctx context.Context, filter *Filter) ([]*Category, error)
	GetCategoryListWithProducts(ctx context.Context, filter *CategoryWithProductsFilter) ([]CategoryWithProducts, error)
}
