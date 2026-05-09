package bas

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Bas) (*Bas, error)
	Update(ctx context.Context, entity *Bas) (*Bas, error)
	Delete(ctx context.Context, entity *Bas) error

	GetByID(ctx context.Context, id string) (*Bas, error)
	GetList(ctx context.Context, filter *Filter) ([]*Bas, int64, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (uc *usecase) Create(ctx context.Context, entity *Bas) (*Bas, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	exists, err := uc.existsByCode(ctx, entity.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check code existence: %w", err)
	}
	if exists {
		return nil, ErrCodeAlreadyExists
	}

	entity.Code = strings.TrimSpace(entity.Code)
	entity.Bas = strings.TrimSpace(entity.Bas)

	created, err := uc.repo.CreateBas(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create bas: %w", err)
	}

	return created, nil
}

func (uc *usecase) Update(ctx context.Context, entity *Bas) (*Bas, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	existing, err := uc.repo.GetBasByID(ctx, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing bas: %w", err)
	}

	if existing.Code != entity.Code {
		exists, err := uc.existsByCode(ctx, entity.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to check code existence: %w", err)
		}
		if exists {
			return nil, ErrCodeAlreadyExists
		}
	}

	entity.Code = strings.TrimSpace(entity.Code)
	entity.Bas = strings.TrimSpace(entity.Bas)

	if err = uc.repo.UpdateBas(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to update bas: %w", err)
	}

	return entity, nil
}

func (uc *usecase) Delete(ctx context.Context, entity *Bas) error {
	existing, err := uc.repo.GetBasByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get bas: %w", err)
	}

	if err = uc.repo.DeleteBas(ctx, existing); err != nil {
		return fmt.Errorf("failed to delete bas: %w", err)
	}

	return nil
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Bas, error) {
	bas, err := uc.repo.GetBasByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bas: %w", err)
	}

	return bas, nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Bas, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	if filter.Keyword != "" {
		filter.Keyword = strings.TrimSpace(filter.Keyword)
	}

	bass, err := uc.repo.GetBasList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bass: %w", err)
	}

	total, err := uc.repo.CountBas(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bass: %w", err)
	}

	return bass, total, nil
}

func (uc *usecase) existsByCode(ctx context.Context, code string) (bool, error) {
	filter := &Filter{
		Code: strings.TrimSpace(code),
	}

	bass, err := uc.repo.GetBasList(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return len(bass) > 0, nil
}
