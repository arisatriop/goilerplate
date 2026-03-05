package bar

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateBar(ctx context.Context, entities *Bar) (*Bar, error)
	UpdateBar(ctx context.Context, entities *Bar) error
	DeleteBar(ctx context.Context, entities *Bar) error
	BulkCreate(ctx context.Context, entities []*Bar) error

	CountBar(ctx context.Context, filter *Filter) (int64, error)
	GetBarList(ctx context.Context, filter *Filter) ([]*Bar, error)
	GetBarByID(ctx context.Context, id string) (*Bar, error)
}
