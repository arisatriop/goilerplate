package zexamplenew

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateExample(ctx context.Context, entities *ZexampleNew) (*ZexampleNew, error)
	UpdateExample(ctx context.Context, entities *ZexampleNew) error
	DeleteExample(ctx context.Context, entities *ZexampleNew) error
	BulkCreate(ctx context.Context, entities []*ZexampleNew) error

	CountExample(ctx context.Context, filter *Filter) (int64, error)
	GetExampleList(ctx context.Context, filter *Filter) ([]*ZexampleNew, error)
	GetExampleByID(ctx context.Context, id string) (*ZexampleNew, error)
}
