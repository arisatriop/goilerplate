package template

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateTemplate(ctx context.Context, entities *Template) (*Template, error)
	UpdateTemplate(ctx context.Context, entities *Template) error
	DeleteTemplate(ctx context.Context, entities *Template) error
	BulkCreate(ctx context.Context, entities []*Template) error

	CountTemplate(ctx context.Context, filter *Filter) (int64, error)
	GetTemplateList(ctx context.Context, filter *Filter) ([]*Template, error)
	GetTemplateByID(ctx context.Context, id string) (*Template, error)
}
