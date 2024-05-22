package v1

import (
	"goilerplate/api/response"
	"goilerplate/app/entity"
	repository "goilerplate/app/repository/v1"

	"github.com/gofiber/fiber/v2"
)

type IExample interface {
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	FindAll(ctx *fiber.Ctx) ([]*response.Example, error)
	FindById(ctx *fiber.Ctx) (*response.Example, error)
}

type ExampleImpl struct {
	Repository repository.IExample
}

func NewExampleUsecase(repository repository.IExample) IExample {
	return &ExampleImpl{
		Repository: repository,
	}
}

func (u *ExampleImpl) Create(ctx *fiber.Ctx) error {
	example := entity.Example{
		Code:      ctx.FormValue("code"),
		Example:   ctx.FormValue("example"),
		CreatedBy: ctx.Get("x-user"),
	}

	return u.Repository.Create(&example)
}

func (u *ExampleImpl) Update(ctx *fiber.Ctx) error {
	panic("Not implement")
}

func (u *ExampleImpl) Delete(ctx *fiber.Ctx) error {
	panic("Not implement")
}

func (u *ExampleImpl) FindAll(ctx *fiber.Ctx) ([]*response.Example, error) {
	panic("Not implement")
}

func (u *ExampleImpl) FindById(ctx *fiber.Ctx) (*response.Example, error) {
	panic("Not implement")
}
