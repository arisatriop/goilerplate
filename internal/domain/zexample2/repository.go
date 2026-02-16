package zexample2

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateExample(ctx context.Context, entities *Example2) (*Example2, error)
	UpdateExample(ctx context.Context, entities *Example2) error
	DeleteExample(ctx context.Context, entities *Example2) error
	BulkCreate(ctx context.Context, entities []*Example2) error

	CountExample(ctx context.Context, filter *Filter) (int64, error)
	GetExampleList(ctx context.Context, filter *Filter) ([]*Example2, error)
	GetExampleByID(ctx context.Context, id string) (*Example2, error)
}
