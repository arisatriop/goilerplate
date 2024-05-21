package v1

import (
	"goilerplate/api/request"
	usecase "goilerplate/src/usecase/v1"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type IExample interface {
	Create() fiber.Handler
	Update() fiber.Handler
	Delete() fiber.Handler
	FindAll() fiber.Handler
	FindById() fiber.Handler
}

type ExampleImpl struct {
	Validator *validator.Validate
	Request   request.IExample
	Usecase   usecase.IExample
}

func NewExampleHandler(validator *validator.Validate, request request.IExample, usecase usecase.IExample) IExample {
	return &ExampleImpl{
		Validator: validator,
		Request:   request,
		Usecase:   usecase,
	}
}

func (h *ExampleImpl) Create() fiber.Handler {
	panic("Not implement")
}

func (h *ExampleImpl) Update() fiber.Handler {
	panic("Not implement")
}

func (h *ExampleImpl) Delete() fiber.Handler {
	panic("Not implement")
}

func (h *ExampleImpl) FindAll() fiber.Handler {
	panic("Not implement")
}

func (h *ExampleImpl) FindById() fiber.Handler {
	panic("Not implement")
}
