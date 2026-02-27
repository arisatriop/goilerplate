package zexample

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateExample(ctx context.Context, entities *Zexample) (*Zexample, error)
	UpdateExample(ctx context.Context, entities *Zexample) error
	DeleteExample(ctx context.Context, entities *Zexample) error
	BulkCreate(ctx context.Context, entities []*Zexample) error

	CountExample(ctx context.Context, filter *Filter) (int64, error)
	GetExampleList(ctx context.Context, filter *Filter) ([]*Zexample, error)
	GetExampleByID(ctx context.Context, id string) (*Zexample, error)
}
