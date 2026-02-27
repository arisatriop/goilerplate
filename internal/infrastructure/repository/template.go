package repository

import (
	"context"
	"goilerplate/internal/domain/template"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type templateRepo struct {
	db *gorm.DB
}

func NewTemplate(db *gorm.DB) template.Repository {
	return &templateRepo{
		db: db,
	}
}

func (r *templateRepo) WithTx(ctx context.Context) template.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &templateRepo{db: tx}
	}
	return r
}

func (r *templateRepo) CreateTemplate(ctx context.Context, entity *template.Template) (*template.Template, error) {
	panic("Implement me")
}

func (r *templateRepo) UpdateTemplate(ctx context.Context, entity *template.Template) error {
	panic("Implement me")
}

func (r *templateRepo) DeleteTemplate(ctx context.Context, entity *template.Template) error {
	panic("Implement me")
}

func (r *templateRepo) GetTemplateByID(ctx context.Context, id string) (*template.Template, error) {
	panic("Implement me")
}

func (r *templateRepo) GetTemplateList(ctx context.Context, filter *template.Filter) ([]*template.Template, error) {
	panic("Implement me")
}

func (r *templateRepo) CountTemplate(ctx context.Context, filter *template.Filter) (int64, error) {
	panic("Implement me")
}

func (r *templateRepo) BulkCreate(ctx context.Context, entities []*template.Template) error {
	panic("Implement me")
}
