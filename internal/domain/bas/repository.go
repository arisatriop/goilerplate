package bas

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateBas(ctx context.Context, entity *Bas) (*Bas, error)
	UpdateBas(ctx context.Context, entity *Bas) error
	DeleteBas(ctx context.Context, entity *Bas) error
	BulkCreate(ctx context.Context, entities []*Bas) error

	CountBas(ctx context.Context, filter *Filter) (int64, error)
	GetBasList(ctx context.Context, filter *Filter) ([]*Bas, error)
	GetBasByID(ctx context.Context, id string) (*Bas, error)
}
