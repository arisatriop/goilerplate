package foo

import (
	"context"

	"github.com/google/uuid"
)

type Usecase interface {
	Create(ctx context.Context, foo *Foo) error
	Update(ctx context.Context, foo *Foo) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	GetList(ctx context.Context, fileter *Filter) ([]Foo, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Foo, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) Create(ctx context.Context, foo *Foo) error {
	panic("Implement me")
}

func (uc *usecase) Update(ctx context.Context, foo *Foo) error {
	panic("Implement me")
}

func (uc *usecase) Delete(ctx context.Context, id uuid.UUID) error {
	panic("Implement me")
}

func (uc *usecase) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("Implement me")
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]Foo, int64, error) {
	panic("Implement me")
}

func (uc *usecase) GetByID(ctx context.Context, id uuid.UUID) (*Foo, error) {
	panic("Implement me")
}
