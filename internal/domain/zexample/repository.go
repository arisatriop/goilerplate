package zexample

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateExample(ctx context.Context, entities *Example) (*Example, error)
	UpdateExample(ctx context.Context, entities *Example) error
	DeleteExample(ctx context.Context, entities *Example) error
	BulkCreate(ctx context.Context, entities []*Example) error

	CountExample(ctx context.Context, filter *Filter) (int64, error)
	GetExampleList(ctx context.Context, filter *Filter) ([]*Example, error)
	GetExampleByID(ctx context.Context, id string) (*Example, error)
}
