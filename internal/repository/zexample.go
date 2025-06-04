package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IExampleRepository interface {
	Create(tx *gorm.DB, example *entity.Example) error
	// FindAll(ctx context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error)
	// FindByID(db *gorm.DB, id string) (*entity.Example, error)
	// Update(tx *gorm.Tx, example *entity.Example) error
	// Delete(tx *gorm.Tx, id string) error
}

type ExampleRepository struct {
	Log *logrus.Logger
}

func NewExampleRepository(log *logrus.Logger) IExampleRepository {
	return &ExampleRepository{
		Log: log,
	}
}

func (r *ExampleRepository) Create(tx *gorm.DB, example *entity.Example) error {
	if err := tx.Create(example).Error; err != nil {
		r.Log.Errorf("failed to create example: %v", err)
		return err
	}
	return nil
}

// func (r *ExampleRepository) FindAll(ctx context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error) {
// 	var examples []entity.Example

// 	query := db.Where("deleted_at IS NULL")
// 	if req.Param.Keyword != "" {
// 		query = query.Where("name LIKE ?", "%"+req.Param.Keyword+"%")
// 	}
// 	if req.OtherTableID != "" {
// 		query = query.Where("other_table_id = ?", req.OtherTableID)
// 	}
// 	if req.Param.Offset > 0 {
// 		query = query.Offset(req.Param.Offset)
// 	}
// 	if req.Param.Limit > 0 {
// 		query = query.Limit(req.Param.Limit)
// 	}

// 	if err := query.Find(&examples).Error; err != nil {
// 		return nil, fmt.Errorf("failed to find examples: %w", err)
// 	}

// 	return examples, nil
// }
