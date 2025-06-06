package repository

import (
	"fmt"
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model/zexample"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IExampleRepository interface {
	Create(db *gorm.DB, example *entity.Example) error
	Update(db *gorm.DB, example *entity.Example) error
	Delete(db *gorm.DB, example *entity.Example) error
	GetAll(db *gorm.DB, req *zexample.GetRequest) error
	GetByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error)
	// FindAll(cdb context.Context, db *gorm.DB, req *model.ExampleGetRequest) ([]entity.Example, error)
}

type ExampleRepository struct {
	Log *logrus.Logger
}

func NewExampleRepository(log *logrus.Logger) IExampleRepository {
	return &ExampleRepository{
		Log: log,
	}
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

func (r *ExampleRepository) Delete(db *gorm.DB, example *entity.Example) error {
	fmt.Printf("Example with ID %s deleted successfully\n", example.DeletedAt)
	fmt.Printf("Example with ID %s deleted successfully\n", example.DeletedBy)

	fmt.Println()
	if err := db.Model(example).UpdateColumns(map[string]any{
		"deleted_at": example.DeletedAt,
		"deleted_by": example.DeletedBy,
	}).Error; err != nil {
		r.Log.Errorf("failed to delete example: %v", err)
		return err
	}

	return nil
}

func (r *ExampleRepository) GetAll(db *gorm.DB, req *zexample.GetRequest) error {
	query := db.Model(&entity.Example{}).Where("deleted_at IS NULL")

	if req.FieldID != "" {
		/**
		 * ! Implement your own filter
		 * ? Example: You want to filter by a specific field ID
		 * *  you can uncomment the following lines
		 */

		// id, err := uuid.Parse(req.FieldID)
		// if err != nil {
		// 	return fmt.Errorf("invalid UUID format for field_id: %w", err)
		// }
		// query = query.Where("field_id = ?", id)
	}

	if req.Keyword != "" {
		query = query.Where("varchar_not_null = ?", "%"+req.Keyword+"%")
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if err := query.Find(&[]entity.Example{}).Error; err != nil {
		r.Log.Errorf("failed to get examples: %v", err)
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

func (r *ExampleRepository) GetByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error) {
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
