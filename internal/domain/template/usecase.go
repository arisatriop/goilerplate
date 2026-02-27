package template

import (
	"context"
)

type Usecase interface {
	Create(ctx context.Context, entity *Template) error
	Update(ctx context.Context, entity *Template) error
	Delete(ctx context.Context, entity *Template) error

	GetByID(ctx context.Context, id string) (*Template, error)
	GetList(ctx context.Context, filter *Filter) ([]*Template, int64, error)

	BulkCreate(ctx context.Context, entities []*Template) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, entity *Template) error {
	panic("Implement me")
}

func (uc *usecase) ExistsByCode(ctx context.Context, code string) (bool, error) {
	panic("Implement me")
}

func (uc *usecase) Update(ctx context.Context, entity *Template) error {
	panic("Implement me")
}

func (uc *usecase) Delete(ctx context.Context, entity *Template) error {
	panic("Implement me")
}

func (uc *usecase) GetByID(ctx context.Context, id string) (*Template, error) {
	panic("Implement me")
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Template, int64, error) {
	panic("Implement me")
}

func (uc *usecase) Count(ctx context.Context, filter *Filter) (int64, error) {
	panic("Implement me")
}

func (uc *usecase) BulkCreate(ctx context.Context, entities []*Template) error {
	panic("Implement me")
}
