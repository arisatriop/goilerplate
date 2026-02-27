package zexamplenew

import "context"

type Usecase interface {
	Create(ctx context.Context, entity *ZexampleNew) error
	Update(ctx context.Context, entity *ZexampleNew) error
	Delete(ctx context.Context, entity *ZexampleNew) error

	GetByID(ctx context.Context, id string) (*ZexampleNew, error)
	GetList(ctx context.Context, filter *Filter) ([]*ZexampleNew, int64, error)

	BulkCreate(ctx context.Context, entities []*ZexampleNew) error
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Create(ctx context.Context, entity *ZexampleNew) error {
	panic("Implemet me")
}

func (u *usecase) Update(ctx context.Context, entity *ZexampleNew) error {
	panic("Implemet me")
}

func (u *usecase) Delete(ctx context.Context, entity *ZexampleNew) error {
	panic("Implemet me")
}

func (u *usecase) GetByID(ctx context.Context, id string) (*ZexampleNew, error) {
	panic("Implemet me")
}

func (u *usecase) GetList(ctx context.Context, filter *Filter) ([]*ZexampleNew, int64, error) {
	panic("Implemet me")
}

func (u *usecase) BulkCreate(ctx context.Context, entities []*ZexampleNew) error {
	panic("Implemet me")
}
