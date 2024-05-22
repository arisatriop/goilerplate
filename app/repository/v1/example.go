package v1

import (
	"context"
	"fmt"
	"goilerplate/api/request"
	"goilerplate/app/entity"
	"goilerplate/config"

	"github.com/jackc/pgx/v5"
)

type IExample interface {
	Create(example *entity.Example) error
	Update(id int64, example *entity.Example) error
	Delete(id int64, example *entity.Example) error
	FindAll(payload *request.ExampleReadPayload) ([]*entity.Example, error)
	FindById(id int64) (*entity.Example, error)
}

type ExampleImpl struct {
	Con *config.Con
}

func NewExampleImpl(con *config.Con) IExample {
	return &ExampleImpl{
		Con: con,
	}
}

func (r *ExampleImpl) Create(example *entity.Example) error {
	if _, err := r.Con.Db.Exec(context.Background(), `
		insert into example (
			code, 
			example, 
			created_by
		) values ($1, $2, $3)`,
		example.Code,
		example.Example,
		example.CreatedBy,
	); err != nil {
		return fmt.Errorf("repository (create example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) Update(id int64, example *entity.Example) error {
	if _, err := r.Con.Db.Exec(context.Background(), `
		update example set 
			code = $1,
			example = $2,
			updated_at = $3,
			updated_by = $4
		where id = $5`,
		example.Code,
		example.Example,
		example.UpdatedAt,
		example.UpdatedBy,
		id,
	); err != nil {
		return fmt.Errorf("repository (update example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) Delete(id int64, example *entity.Example) error {
	stmt, err := r.Con.Db.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}

	if _, err = stmt.Exec(context.Background(), `
		update example set
			deleted_at = $1,
			deleted_by = $2
		where id = $3`,
		example.DeletedAt,
		example.DeletedBy,
		example.Id,
	); err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) FindAll(payload *request.ExampleReadPayload) ([]*entity.Example, error) {
	panic("Not implement")
}

func (r *ExampleImpl) FindById(id int64) (*entity.Example, error) {
	var exp entity.Example

	row := r.Con.Db.QueryRow(context.Background(), `
		select 
		id,
		code,
		example,
		created_at,
		created_by,
		updated_at,
		updated_by,
		deleted_at,
		deleted_by,
		uuid
		from example 
		where id = $1`, id,
	)

	err := row.Scan(
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
	)
	if err != nil && err != pgx.ErrNoRows {
		return &exp, fmt.Errorf("repository (find by id example): %v", err)
	}

	return &exp, err
}
