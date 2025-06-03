package repository

import (
	"context"
	"fmt"
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"

	"gorm.io/gorm"
)

type ExampleRepository struct {
}

type IExampleRepository interface {
	FindAll(ctx context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error)
	// FindByID(db *gorm.DB, id string) (*entity.Example, error)
	// Create(tx *gorm.Tx, example *entity.Example) error
	// Update(tx *gorm.Tx, example *entity.Example) error
	// Delete(tx *gorm.Tx, id string) error
}

func NewExampleRepository() IExampleRepository {
	return &ExampleRepository{}
}

func (r *ExampleRepository) FindAll(ctx context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error) {
	var examples []entity.Example

	query := db.Where("deleted_at IS NULL")
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.SomeID != "" {
		query = query.Where("some_id = ?", req.SomeID)
	}
	if req.Paginate.Offset > 0 {
		query = query.Offset(req.Paginate.Offset)
	}
	if req.Paginate.Limit > 0 {
		query = query.Limit(req.Paginate.Limit)
	}

	if err := query.Find(&examples).Error; err != nil {
		return nil, fmt.Errorf("failed to find examples: %w", err)
	}

	return examples, nil
}
