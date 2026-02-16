package category

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

type Usecase interface {
	GetList(ctx context.Context, filter *Filter) ([]*Category, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID, filter *Filter) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) SoftDelete(ctx context.Context, id uuid.UUID, filter *Filter) error {
	category, err := uc.repo.GetCategoryByID(ctx, id, filter)
	if err != nil {
		return fmt.Errorf("failed to get category: %v", err)
	}
	if category == nil {
		return utils.Error(http.StatusNotFound, "category not found")
	}

	if err := uc.repo.SoftDeleteCategory(ctx, category.ID, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
		return fmt.Errorf("failed to soft delete category: %v", err)
	}

	return nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Category, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	categories, err := uc.repo.GetCategoryList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get category: %v", err)
	}

	total, err := uc.repo.CountCategories(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %v", err)
	}

	return categories, total, nil
}
