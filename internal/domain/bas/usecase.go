package bas

import (
	"context"
	"strings"
)

type Usecase interface {
	Create(ctx context.Context, entity *Bas) (*Bas, error)
	Update(ctx context.Context, entity *Bas) (*Bas, error)
	Delete(ctx context.Context, entity *Bas) error

	GetByID(ctx context.Context, id string) (*Bas, error)
	GetList(ctx context.Context, filter *Filter) ([]*Bas, int64, error)

	BulkCreate(ctx context.Context, entities []*Bas) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Bas) (*Bas, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Name = strings.TrimSpace(entity.Name)

	return &Bas{
		ID:   "hardcoded-id-001",
		Code: entity.Code,
		Name: entity.Name,
	}, nil
}

func (uc *usecase) Update(ctx context.Context, entity *Bas) (*Bas, error) {
	if err := entity.validate(); err != nil {
		return nil, err
	}

	entity.Code = strings.ToUpper(strings.TrimSpace(entity.Code))
	entity.Name = strings.TrimSpace(entity.Name)

	return &Bas{
		ID:   entity.ID,
		Code: entity.Code,
		Name: entity.Name,
	}, nil
}

func (uc *usecase) Delete(ctx context.Context, entity *Bas) error {
	return nil
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Bas, error) {
	return &Bas{
		ID:   id,
		Code: "BAS001",
		Name: "Sample Bas",
	}, nil
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Bas, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	samples := []*Bas{
		{ID: "hardcoded-id-001", Code: "BAS001", Name: "Sample Bas 1"},
		{ID: "hardcoded-id-002", Code: "BAS002", Name: "Sample Bas 2"},
	}

	return samples, int64(len(samples)), nil
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Bas) error {
	for i, entity := range entities {
		if err := entity.validate(); err != nil {
			return err
		}
		entities[i].Code = strings.ToUpper(strings.TrimSpace(entity.Code))
		entities[i].Name = strings.TrimSpace(entity.Name)
	}
	return nil
}
