package product

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

type Usecase interface {
	GetMenuProductList(ctx context.Context, filter *Filter) ([]*Product, int64, error)
	UpdateProduct(ctx context.Context, prod *Product) error
	UpdateToggleAvailable(ctx context.Context, productID uuid.UUID) (bool, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) GetMenuProductList(ctx context.Context, filter *Filter) ([]*Product, int64, error) {
	products, err := uc.repo.GetMenuProductList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get product list: %w", err)
	}

	total, err := uc.repo.CountMenuProductList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count product list: %w", err)
	}

	return products, total, nil
}

func (uc *usecase) UpdateProduct(ctx context.Context, prod *Product) error {
	// Verify product exists and belongs to the store
	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	existingProd, err := uc.repo.GetProductByID(ctx, prod.ID, &Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if existingProd == nil {
		return utils.Error(http.StatusNotFound, ErrProductNotFound)
	}

	existingProd.Name = prod.Name
	existingProd.Description = prod.Description
	existingProd.Price = prod.Price
	existingProd.Images = prod.Images
	if err := uc.repo.UpdateProduct(ctx, existingProd); err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

func (uc *usecase) UpdateToggleAvailable(ctx context.Context, productID uuid.UUID) (bool, error) {
	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	prod, err := uc.repo.GetProductByID(ctx, productID, &Filter{StoreID: &storeID})
	if err != nil {
		return false, fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return false, utils.Error(http.StatusNotFound, ErrProductNotFound)
	}

	prod.IsAvailable = !prod.IsAvailable
	if err := uc.repo.UpdateToggleAvailable(ctx, prod.ID, prod.IsAvailable, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
		return false, fmt.Errorf("failed to update product availability: %v", err)
	}

	return prod.IsAvailable, nil

}
