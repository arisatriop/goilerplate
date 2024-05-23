package v1

import (
	"context"
	"fmt"
	"goilerplate/api/request"
	"goilerplate/app/entity"
	"goilerplate/config"
	"time"

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

func NewExampleRepository(con *config.Con) IExample {
	return &ExampleImpl{
		Con: con,
	}
}

func (r *ExampleImpl) Create(example *entity.Example) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := r.Con.Db.Exec(ctx, `
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := r.Con.Db.Exec(ctx, `
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := r.Con.Db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}
	defer conn.Release()

	stmt := "deleteExampleById"
	if _, err = conn.Conn().Prepare(context.Background(), stmt, `
		update example set
			deleted_at = $1,
			deleted_by = $2
		where id = $3`,
	); err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}

	if _, err = conn.Conn().Exec(context.Background(),
		stmt,
		example.DeletedAt,
		example.DeletedBy,
		example.Id,
	); err != nil {
		return fmt.Errorf("repository (delete example): %v", err)
	}

	return nil
}

func (r *ExampleImpl) FindAll(payload *request.ExampleReadPayload) ([]*entity.Example, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exps []*entity.Example

	rows, err := r.Con.Db.Query(ctx, getSqlFindAll(payload), getArgsFindAll(payload))
	if err != nil {
		return nil, fmt.Errorf("repository (find all example): %v", err)
	}

	return exps, nil
}

func (r *ExampleImpl) FindById(id int64) (*entity.Example, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exp entity.Example

	conn, err := r.Con.Db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository (find by id example): %v", err)
	}
	defer conn.Release()

	stmt := "findExampleById"
	if _, err = conn.Conn().Prepare(context.Background(), stmt, `
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
		where id = $1
		and deleted_at is null`,
	); err != nil {
		return nil, fmt.Errorf("repository (find by id example): %v", err)
	}

	if err = conn.QueryRow(context.Background(), stmt, id).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repository (find by id example): %v", err)
	}

	return &exp, err
}

// TODO
// Make this singleton
func GetSqlFindAllExample(payload *request.ExampleReadPayload) string {
	sql := `
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
		where deleted_at is null
	`

	if payload.Search != "" {
		sql += " and ()"
	}

}
