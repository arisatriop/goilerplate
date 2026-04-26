package bar

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Bar) (*Bar, error)
	Update(ctx context.Context, entity *Bar) error
	Delete(ctx context.Context, entity *Bar) error

	GetByID(ctx context.Context, id string) (*Bar, error)
	GetList(ctx context.Context, filter *Filter) ([]*Bar, int64, error)

	BulkCreate(ctx context.Context, entities []*Bar) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Bar) (*Bar, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	exists, err := uc.ExistsByCode(ctx, entity.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check code existence: %w", err)
	}
	if exists {
		return nil, ErrCodeAlreadyExists
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Bar = strings.TrimSpace(entity.Bar)

	created, err := uc.repo.CreateBar(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create bar: %w", err)
	}

	return created, nil
}

func (uc *usecase) ExistsByCode(ctx context.Context, code string) (bool, error) {
	filter := &Filter{
		Code: strings.ToUpper(strings.TrimSpace(code)),
	}

	bars, err := uc.repo.GetBarList(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return len(bars) > 0, nil
}

func (uc *usecase) Update(ctx context.Context, entity *Bar) error {
	if err := entity.validate(); err != nil {
		return err
	}

	existing, err := uc.repo.GetBarByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing bar: %w", err)
	}

	if existing.Code != entity.Code {
		exists, err := uc.ExistsByCode(ctx, entity.Code)
		if err != nil {
			return fmt.Errorf("failed to check code existence: %w", err)
		}
		if exists {
			return ErrCodeAlreadyExists
		}
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Bar = strings.TrimSpace(entity.Bar)

	if err = uc.repo.UpdateBar(ctx, entity); err != nil {
		return fmt.Errorf("failed to update bar: %w", err)
	}

	return nil
}

func (uc *usecase) Delete(ctx context.Context, entity *Bar) error {
	existing, err := uc.repo.GetBarByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get bar: %w", err)
	}

	if err = uc.repo.DeleteBar(ctx, existing); err != nil {
		return fmt.Errorf("failed to delete bar: %w", err)
	}

	return nil
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Bar, error) {
	bar, err := uc.repo.GetBarByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bar: %w", err)
	}

	return bar, nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Bar, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	if filter.Keyword != "" {
		filter.Keyword = strings.TrimSpace(filter.Keyword)
	}

	bars, err := uc.repo.GetBarList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bars: %w", err)
	}

	total, err := uc.repo.CountBar(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bars: %w", err)
	}

	return bars, total, nil
}

func (uc *usecase) Count(ctx context.Context, filter *Filter) (int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	count, err := uc.repo.CountBar(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count bars: %w", err)
	}

	return count, nil
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Bar) error {
	codes := make(map[string]bool)
	for i, entity := range entities {
		if err := entity.validate(); err != nil {
			return fmt.Errorf("validation failed for entity %d: %w", i, err)
		}

		code := strings.ToUpper(strings.TrimSpace(entity.Code))
		if codes[code] {
			return fmt.Errorf("duplicate code '%s' in batch", code)
		}
		codes[code] = true

		entity.Code = code
		entity.Bar = strings.TrimSpace(entity.Bar)
	}

	// Check existing codes in database
	for code := range codes {
		exists, err := uc.ExistsByCode(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to check code existence for '%s': %w", code, err)
		}
		if exists {
			return fmt.Errorf("code '%s' already exists", code)
		}
	}

	// Bulk create
	if err := uc.repo.BulkCreate(ctx, entities); err != nil {
		return fmt.Errorf("failed to bulk create bars: %w", err)
	}

	return nil
}
