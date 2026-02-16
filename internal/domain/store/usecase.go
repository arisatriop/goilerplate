package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Usecase interface {
	GetInfo(ctx context.Context, storeID string) (*Store, error)
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (u *usecase) GetInfo(ctx context.Context, storeID string) (*Store, error) {
	id, err := uuid.Parse(storeID)
	if err != nil {
		return nil, err
	}

	store, err := u.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get store by id: %w", err)
	}

	return store, nil
}
