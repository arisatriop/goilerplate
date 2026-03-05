package handler

import (
	"goilerplate/internal/domain/foo"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Foo struct {
	Validator *validator.Validate
	Usecase   foo.Usecase
}

func NewFoo(validator *validator.Validate, usecase foo.Usecase) *Foo {
	return &Foo{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *Foo) Create(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Foo) Update(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Foo) Delete(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Foo) List(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Foo) Get(ctx *fiber.Ctx) error {
	panic("Implement me")
}
