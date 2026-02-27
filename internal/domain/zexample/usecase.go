package zexample

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Zexample) error
	Update(ctx context.Context, entity *Zexample) error
	Delete(ctx context.Context, entity *Zexample) error

	GetByID(ctx context.Context, id string) (*Zexample, error)
	GetList(ctx context.Context, filter *Filter) ([]*Zexample, int64, error)

	// BulkCreate(ctx context.Context, entities []*Example) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Zexample) error {
	if err := entity.validate(); err != nil {
		return err
	}

	exists, err := uc.ExistsByCode(ctx, entity.Code)
	if err != nil {
		return fmt.Errorf("failed to check code existence: %w", err)
	}
	if exists {
		return ErrCodeAlreadyExists
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Example = strings.TrimSpace(entity.Example)

	_, err = uc.repo.CreateExample(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to create example: %w", err)
	}

	return nil
}

func (uc *usecase) ExistsByCode(ctx context.Context, code string) (bool, error) {
	filter := &Filter{
		Code: strings.ToUpper(strings.TrimSpace(code)),
	}

	examples, err := uc.repo.GetExampleList(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return len(examples) > 0, nil
}

func (uc *usecase) Update(ctx context.Context, entity *Zexample) error {
	if err := entity.validate(); err != nil {
		return err
	}

	existing, err := uc.repo.GetExampleByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing example: %w", err)
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
	entity.Example = strings.TrimSpace(entity.Example)

	if err = uc.repo.UpdateExample(ctx, entity); err != nil {
		return fmt.Errorf("failed to update example: %w", err)
	}

	return nil
}

func (uc *usecase) Delete(ctx context.Context, entity *Zexample) error {
	existing, err := uc.repo.GetExampleByID(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get example: %w", err)
	}

	if err = uc.repo.DeleteExample(ctx, existing); err != nil {
		return fmt.Errorf("failed to delete example: %w", err)
	}

	return nil
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Zexample, error) {
	example, err := uc.repo.GetExampleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get example: %w", err)
	}

	return example, nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Zexample, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	if filter.Keyword != "" {
		filter.Keyword = strings.TrimSpace(filter.Keyword)
	}

	examples, err := uc.repo.GetExampleList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get examples: %w", err)
	}

	total, err := uc.repo.CountExample(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count examples: %w", err)
	}

	return examples, total, nil
}

func (uc *usecase) Count(ctx context.Context, filter *Filter) (int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	count, err := uc.repo.CountExample(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count examples: %w", err)
	}

	return count, nil
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Zexample) error {
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
		entity.Example = strings.TrimSpace(entity.Example)
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
		return fmt.Errorf("failed to bulk create examples: %w", err)
	}

	return nil
}
