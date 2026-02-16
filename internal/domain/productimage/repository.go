package productimage

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateProductImage(ctx context.Context, productImage *ProductImage) error
	UpdateProductImage(ctx context.Context, productImage *ProductImage) error

	DeleteProductImageByID(ctx context.Context, id uuid.UUID) error
	DeleteProductImagesByProductID(ctx context.Context, productID uuid.UUID) error

	ResetPrimaryImage(ctx context.Context, producID uuid.UUID) error

	GetProductImagesByProductID(ctx context.Context, productID uuid.UUID) ([]ProductImage, error)
	GetProductImageByStoreID(ctx context.Context, storeID uuid.UUID) ([]ProductImage, error)
	GetProductImageByID(ctx context.Context, id uuid.UUID) (*ProductImage, error)
}
