package bazs

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Bazs) (*Bazs, error)
	Update(ctx context.Context, entity *Bazs) (*Bazs, error)
	Delete(ctx context.Context, entity *Bazs) error

	GetByID(ctx context.Context, id string) (*Bazs, error)
	GetList(ctx context.Context, filter *Filter) ([]*Bazs, int64, error)

	BulkCreate(ctx context.Context, entities []*Bazs) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Bazs) (*Bazs, error) {
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
	entity.Name = strings.TrimSpace(entity.Name)

	created, err := uc.repo.CreateBazs(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create bazs: %w", err)
	}

	return created, nil
}

func (uc *usecase) ExistsByCode(ctx context.Context, code string) (bool, error) {
	filter := &Filter{
		Code: strings.ToUpper(strings.TrimSpace(code)),
	}

	items, err := uc.repo.GetBazsList(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return len(items) > 0, nil
}

func (uc *usecase) Update(ctx context.Context, entity *Bazs) (*Bazs, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	existing, err := uc.repo.GetBazsByID(ctx, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing bazs: %w", err)
	}

	if existing.Code != entity.Code {
		exists, err := uc.ExistsByCode(ctx, entity.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to check code existence: %w", err)
		}
		if exists {
			return nil, ErrCodeAlreadyExists
		}
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Name = strings.TrimSpace(entity.Name)

	if err = uc.repo.UpdateBazs(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to update bazs: %w", err)
	}

	return entity, nil
}

func (uc *usecase) Delete(ctx context.Context, entity *Bazs) error {
	existing, err := uc.repo.GetBazsByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get bazs: %w", err)
	}

	if err = uc.repo.DeleteBazs(ctx, existing); err != nil {
		return fmt.Errorf("failed to delete bazs: %w", err)
	}

	return nil
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Bazs, error) {
	item, err := uc.repo.GetBazsByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bazs: %w", err)
	}

	return item, nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Bazs, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	if filter.Keyword != "" {
		filter.Keyword = strings.TrimSpace(filter.Keyword)
	}

	items, err := uc.repo.GetBazsList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bazs list: %w", err)
	}

	total, err := uc.repo.CountBazs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bazs: %w", err)
	}

	return items, total, nil
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Bazs) error {
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
		entity.Name = strings.TrimSpace(entity.Name)
	}

	for code := range codes {
		exists, err := uc.ExistsByCode(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to check code existence for '%s': %w", code, err)
		}
		if exists {
			return fmt.Errorf("code '%s' already exists", code)
		}
	}

	if err := uc.repo.BulkCreate(ctx, entities); err != nil {
		return fmt.Errorf("failed to bulk create bazs: %w", err)
	}

	return nil
}
