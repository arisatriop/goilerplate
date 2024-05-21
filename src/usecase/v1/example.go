package v1

import (
	"goilerplate/api/response"
	repository "goilerplate/src/repository/v1"

	"github.com/gofiber/fiber/v2"
)

type IExample interface {
	Create(request *fiber.Request) error
	Update(request *fiber.Request) error
	Delete(id int64) error
	FindAll() ([]response.Example, error)
	FindById(request *fiber.Request) (*response.Example, error)
}

type ExampleImpl struct {
	Repository repository.IExample
}

func NewExampleUsecase(repository repository.IExample) IExample {
	return &ExampleImpl{
		Repository: repository,
	}
}

func (u *ExampleImpl) Create(request *fiber.Request) error {
	panic("Not implement")
}

func (u *ExampleImpl) Update(request *fiber.Request) error {
	panic("Not implement")
}

func (u *ExampleImpl) Delete(id int64) error {
	panic("Not implement")
}

func (u *ExampleImpl) FindAll() ([]response.Example, error) {
	panic("Not implement")
}

func (u *ExampleImpl) FindById(request *fiber.Request) (*response.Example, error) {
	panic("Not implement")
}
