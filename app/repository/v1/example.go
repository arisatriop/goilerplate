package v1

import (
	"context"
	"fmt"
	"goilerplate/api/request"
	"goilerplate/app/entity"
	"time"

	"gorm.io/gorm"
)

type IExample interface {
	Create(tx *gorm.DB, example *entity.Example) error
	Update(tx *gorm.DB, id string, example *entity.Example) error
	Delete(tx *gorm.DB, id string, example *entity.Example) error
	FindAll(db *gorm.DB, payload *request.ExampleReadPayload) ([]*entity.Example, error)
	FindById(db *gorm.DB, id int64) (*entity.Example, error)
}

type ExampleImpl struct{}

func NewExampleRepository() IExample {
	return &ExampleImpl{}
}

func (r *ExampleImpl) Create(tx *gorm.DB, example *entity.Example) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := tx.WithContext(ctx).Table("example").Create(&example).Error; err != nil {
		return fmt.Errorf("repository (create example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) Update(tx *gorm.DB, id string, example *entity.Example) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := tx.WithContext(ctx).Table("example").Where("id = ?", id).Updates(&example).Error; err != nil {
		return fmt.Errorf("repository (update example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) Delete(tx *gorm.DB, id string, example *entity.Example) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := tx.WithContext(ctx).Table("example").Where("id = ?", id).Updates(&example).Error; err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) FindAll(db *gorm.DB, payload *request.ExampleReadPayload) ([]*entity.Example, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exps []*entity.Example

	where := "deleted_at is null"

	if payload.Search != "" {
		where += " and (code ILIKE '%' || " + payload.Search + " || '%' or example ILIKE '%' || " + payload.Search + " || '%')"
	}
	if payload.Limit != "" {
		where += " limit " + payload.Limit
	}
	if payload.Offset != "" {
		where += " offset " + payload.Offset
	}

	err := db.WithContext(ctx).Table("example").Where(where).Find(&exps).Error
	if err != nil {
		return exps, fmt.Errorf("repository (find all example): %v", err)
	}

	return exps, err
}

func (r *ExampleImpl) FindById(db *gorm.DB, id int64) (*entity.Example, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exp entity.Example

	err := db.WithContext(ctx).Table("example").Where("id = ? and deleted_at is null", id).First(&exp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return &exp, fmt.Errorf("repository (find by id example): %v", err)
	}

	return &exp, err
}
