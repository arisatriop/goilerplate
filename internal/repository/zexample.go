package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IExampleRepository interface {
	FindByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error)
	Create(db *gorm.DB, example *entity.Example) error
	Update(db *gorm.DB, example *entity.Example) error
	// FindAll(cdb context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error)
	// Delete(db *gorm.db, id string) error
}

type ExampleRepository struct {
	Log *logrus.Logger
}

func NewExampleRepository(log *logrus.Logger) IExampleRepository {
	return &ExampleRepository{
		Log: log,
	}
}

func (r *ExampleRepository) FindByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error) {
	example := &entity.Example{}
	if err := db.Where("id = ? and deleted_at is null", id).First(example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to find example by ID %s: %v", id, err)
		return nil, err
	}
	return example, nil
}

func (r *ExampleRepository) Create(db *gorm.DB, example *entity.Example) error {
	if err := db.Create(example).Error; err != nil {
		r.Log.Errorf("failed to create example: %v", err)
		return err
	}
	return nil
}

func (r *ExampleRepository) Update(db *gorm.DB, example *entity.Example) error {
	if err := db.Save(example).Error; err != nil {
		r.Log.Errorf("failed to update example: %v", err)
		return err
	}
	return nil
}

// func (r *ExampleRepository) FindAll(cdb context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error) {
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
