package v2

import (
	"context"
	"fmt"
	"goilerplate/api/request"
	"goilerplate/app/entity"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

type IExample interface {
	Create(tx *gorm.DB, example *entity.Example) error
	Update(tx *gorm.DB, id string, example *entity.Example) error
	Delete(tx *gorm.DB, id string, example *entity.Example) error
	FindAll(db *pgxpool.Pool, payload *request.ExampleReadPayload) ([]*entity.Example, error)
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

func (r *ExampleImpl) FindAll(db *pgxpool.Pool, payload *request.ExampleReadPayload) ([]*entity.Example, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exps []*entity.Example

	rows, err := db.Query(ctx, getSqlFindAllExample(payload), getArgsFindAllExample(payload)...)
	if err != nil {
		return nil, fmt.Errorf("repository (find all example): %v", err)
	}

	for rows.Next() {
		var exp entity.Example
		if err = rows.Scan(
			&exp.Id,
			&exp.Code,
			&exp.Example,
			&exp.CreatedAt,
			&exp.CreatedBy,
			&exp.UpdatedAt,
			&exp.UpdatedBy,
			&exp.DeletedAt,
			&exp.DeletedBy,
			&exp.Uuid,
		); err != nil {
			return nil, fmt.Errorf("repository (find all example): %v", err)
		}
		exps = append(exps, &exp)
	}

	return exps, nil
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

// TODO
// Make this singleton
func getSqlFindAllExample(payload *request.ExampleReadPayload) string {

	sql := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null"

	search, limit, offset := getPlaceHolder(payload)

	if search != "$0" {
		sql += " and (code ilike '%' || $1 || '%' or example ilike '%' || $1 || '%')"
	}

	if limit != "$0" {
		sql += " limit " + limit
	}

	if offset != "$0" {
		sql += " offset " + offset
	}

	return sql
}

func getArgsFindAllExample(payload *request.ExampleReadPayload) []interface{} {

	var args []interface{}

	search, limit, offset := getPlaceHolder(payload)
	if search != "$0" {
		args = append(args, payload.Search)
	}
	if limit != "$0" {
		args = append(args, payload.Limit)
	}
	if offset != "$0" {
		args = append(args, payload.Offset)
	}

	return args
}

func getPlaceHolder(payload *request.ExampleReadPayload) (string, string, string) {
	search := 0
	limit := 0
	offset := 0

	if payload.Search != "" {
		search = 1
	}

	if payload.Limit != "" && payload.Limit != "0" {
		limit = 1
		if search != 0 {
			limit += 1
		}
	}

	if payload.Offset != "" && payload.Offset != "0" {
		offset = 1
		if search != 0 {
			offset += 1
		}
		if limit != 0 {
			offset += 1
		}
	}

	return fmt.Sprintf("$%d", search), fmt.Sprintf("$%d", limit), fmt.Sprintf("$%d", offset)
}
