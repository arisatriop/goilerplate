package bazs

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateBazs(ctx context.Context, entity *Bazs) (*Bazs, error)
	UpdateBazs(ctx context.Context, entity *Bazs) error
	DeleteBazs(ctx context.Context, entity *Bazs) error
	BulkCreate(ctx context.Context, entities []*Bazs) error

	CountBazs(ctx context.Context, filter *Filter) (int64, error)
	GetBazsList(ctx context.Context, filter *Filter) ([]*Bazs, error)
	GetBazsByID(ctx context.Context, id string) (*Bazs, error)
}
