package v1

import (
	"database/sql"
	"fmt"
	"goilerplate/api/request"
	"goilerplate/app/entity"
	repository "goilerplate/app/repository/v1"
	"goilerplate/config"
	"goilerplate/helper"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IExample interface {
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	FindAll(ctx *fiber.Ctx) ([]*entity.Example, error)
	FindById(ctx *fiber.Ctx) (*entity.Example, error)
}

type ExampleImpl struct {
	App        *config.App
	Repository repository.IExample
}

func NewExampleUsecase(app *config.App, repository repository.IExample) IExample {
	return &ExampleImpl{
		App:        app,
		Repository: repository,
	}
}

func (u *ExampleImpl) Create(ctx *fiber.Ctx) error {

	tx := u.App.DB.Gdb.Begin()

	example := entity.Example{
		Id:        helper.GenerateShortUUID(),
		Code:      ctx.FormValue("code"),
		Example:   ctx.FormValue("example"),
		CreatedBy: ctx.Get("x-user"),
	}

	if err := u.Repository.Create(tx, &example); err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (create example): %s", err)
		}
		return fmt.Errorf("usecase (create example): %s", err)
	}

	if err := tx.Commit().Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (create example): %s", err)
		}
		return fmt.Errorf("usecase (create example): %s", err)
	}

	return nil
}

func (u *ExampleImpl) Update(ctx *fiber.Ctx) error {

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("usecase (update example): %s", err)
	}

	example, err := u.Repository.FindById(u.App.DB.Gdb, int64(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return fmt.Errorf("usecase (update example): %s", err)
	}

	example.Code = ctx.FormValue("code")
	example.Example = ctx.FormValue("example")
	example.UpdatedBy = sql.NullString{String: ctx.Get("x-user"), Valid: true}

	tx := u.App.DB.Gdb.Begin()

	err = u.Repository.Update(tx, example.Id, example)
	if err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (update example): %s", err)
		}
		return fmt.Errorf("usecase (update example): %s", err)
	}

	if err := tx.Commit().Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (update example): %s", err)
		}
		return fmt.Errorf("usecase (update example): %s", err)
	}

	return nil
}

func (u *ExampleImpl) Delete(ctx *fiber.Ctx) error {

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("usecase (delete example): %s", err)
	}

	example, err := u.Repository.FindById(u.App.DB.Gdb, int64(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return fmt.Errorf("usecase (delete example): %s", err)
	}

	example.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	example.DeletedBy = sql.NullString{String: ctx.Get("x-user"), Valid: true}

	tx := u.App.DB.Gdb.Begin()

	err = u.Repository.Update(tx, example.Id, example)
	if err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (delete example): %s", err)
		}
		return fmt.Errorf("usecase (delete example): %s", err)
	}

	if err := tx.Commit().Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return fmt.Errorf("usecase (delete example): %s", err)
		}
		return fmt.Errorf("usecase (delete example): %s", err)
	}

	return nil
}

func (u *ExampleImpl) FindAll(ctx *fiber.Ctx) ([]*entity.Example, error) {

	payload := request.ExampleReadPayload{
		Search: ctx.FormValue("search"),
		Limit:  ctx.FormValue("limit"),
		Offset: ctx.FormValue("offset"),
	}

	examples, err := u.Repository.FindAll(u.App.DB.Gdb, &payload)
	if err != nil {
		return nil, fmt.Errorf("usecase (find all example): %s", err)
	}

	return examples, nil
}

func (u *ExampleImpl) FindById(ctx *fiber.Ctx) (*entity.Example, error) {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return nil, fmt.Errorf("usecase (find by id example): %s", err)
	}

	example, err := u.Repository.FindById(u.App.DB.Gdb, int64(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("usecase (find by id example): %s", err)
	}

	return example, nil
}
