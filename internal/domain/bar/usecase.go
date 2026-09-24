package bar

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Bar) (*Bar, error)
	Update(ctx context.Context, entity *Bar) (*Bar, error)
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

// Create stores a new bar. Uniqueness of the code is enforced by the database, across deleted
// bars too; the repository reports a clash as ErrCodeAlreadyExists.
func (uc *usecase) Create(ctx context.Context, entity *Bar) (*Bar, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}
	entity.normalize()

	created, err := uc.repo.CreateBar(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("creating bar: %w", err)
	}

	return created, nil
}

// Update changes a bar's content. The code is the business key and never changes: a request may
// leave it out or repeat it — in any case or spacing — and anything else is ErrCodeImmutable.
func (uc *usecase) Update(ctx context.Context, entity *Bar) (*Bar, error) {
	existing, err := uc.repo.GetBarByID(ctx, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("loading bar: %w", err)
	}
	if entity.Code != "" && normalizeCode(entity.Code) != existing.Code {
		return nil, ErrCodeImmutable
	}
	entity.Code = existing.Code

	if err := entity.validate(); err != nil {
		return nil, err
	}
	entity.normalize()

	if err := uc.repo.UpdateBar(ctx, entity); err != nil {
		return nil, fmt.Errorf("updating bar: %w", err)
	}

	return entity, nil
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

// BulkCreate stores every bar or none. A code repeated inside the batch is the client's
// mistake and is refused before the write; a clash with an existing bar comes back from the
// repository as ErrCodeAlreadyExists.
func (uc *usecase) BulkCreate(ctx context.Context, entities []*Bar) error {
	seen := make(map[string]bool, len(entities))
	for i, entity := range entities {
		if err := entity.validate(); err != nil {
			return fmt.Errorf("validation failed for entity %d: %w", i, err)
		}
		entity.normalize()

		if seen[entity.Code] {
			return ErrDuplicateCodeInBatch
		}
		seen[entity.Code] = true
	}

	if err := uc.repo.BulkCreate(ctx, entities); err != nil {
		return fmt.Errorf("bulk creating bars: %w", err)
	}

	return nil
}
