package repository

import (
	"goilerplate/internal/entity"
	"goilerplate/internal/model/zexample"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ExampleRepository interface {
	Create(db *gorm.DB, example *entity.Example) error
	Update(db *gorm.DB, example *entity.Example) error
	Delete(db *gorm.DB, example *entity.Example) error
	GetAll(db *gorm.DB, req *zexample.GetRequest) ([]entity.Example, int64, error)
	GetByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error)
	Count(db *gorm.DB, req *zexample.GetRequest) (int64, error)

	// BatchInsert(ctx context.Context, db *gorm.DB, examples []entity.Example) error
	// DeleteByFieldID(ctx context.Context, db *gorm.DB, fieldID uuid.UUID) error
	// HardDelete(ctx context.Context, db *gorm.DB, example *entity.Example) error
	// HardDeleteByFieldID(ctx context.Context, db *gorm.DB, fieldID uuid.UUID) error
}

type exampleRepository struct {
	Log *logrus.Logger
}

func NewExampleRepository(log *logrus.Logger) ExampleRepository {
	return &exampleRepository{
		Log: log,
	}
}

func (r *exampleRepository) Create(db *gorm.DB, example *entity.Example) error {
	if err := db.Create(example).Error; err != nil {
		r.Log.Errorf("failed to create example: %v", err)
		return err
	}
	return nil
}

func (r *exampleRepository) Update(db *gorm.DB, example *entity.Example) error {
	if err := db.Save(example).Error; err != nil {
		r.Log.Errorf("failed to update example: %v", err)
		return err
	}
	return nil
}

func (r *exampleRepository) Delete(db *gorm.DB, example *entity.Example) error {
	if err := db.Model(example).UpdateColumns(map[string]any{
		"deleted_at": example.DeletedAt,
		"deleted_by": example.DeletedBy,
	}).Error; err != nil {
		r.Log.Errorf("failed to delete example: %v", err)
		return err
	}
	return nil
}

func (r *exampleRepository) GetAll(db *gorm.DB, req *zexample.GetRequest) ([]entity.Example, int64, error) {
	query := db.Model(&entity.Example{}).Where("deleted_at IS NULL")

	if req.FieldID != "" {
		/**
		 * ! Implement your own filter
		 * ? Example: You want to filter by a specific field ID
		 * *  you can uncomment the following lines
		 */

		// id, err := uuid.Parse(req.FieldID)
		// if err != nil {
		// 	return nil, helper.Error(http.StatusBadRequest, "invalid field_id format", err)
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

	var examples []entity.Example
	if err := query.Find(&examples).Error; err != nil {
		r.Log.Errorf("failed to get examples: %v", err)
		return nil, 0, err
	}

	count, err := r.Count(db, req)
	if err != nil {
		return nil, 0, err
	}

	return examples, count, nil
}

func (r *exampleRepository) GetByID(db *gorm.DB, id uuid.UUID) (*entity.Example, error) {
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

func (r *exampleRepository) Count(db *gorm.DB, req *zexample.GetRequest) (int64, error) {
	query := db.Model(&entity.Example{}).Where("deleted_at IS NULL")

	if req.FieldID != "" {
		/**
		 * ! Implement your own filter
		 * ? Example: You want to filter by a specific field ID
		 * *  you can uncomment the following lines
		 */

		// id, err := uuid.Parse(req.FieldID)
		// if err != nil {
		// 	return 0, helper.Error(http.StatusBadRequest, "invalid field_id format", err)
		// }
		// query = query.Where("field_id = ?", id)
	}
	if req.Keyword != "" {
		query = query.Where("varchar_not_null = ?", "%"+req.Keyword+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.Log.Errorf("failed to count examples: %v", err)
		return 0, err
	}

	return count, nil
}
