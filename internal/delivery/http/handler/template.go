package handler

import (
	"goilerplate/internal/domain/template"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Template struct {
	Validator *validator.Validate
	Usecase   template.Usecase
}

func NewTemplate(validator *validator.Validate, usecase template.Usecase) *Template {
	return &Template{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *Template) Create(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Template) Update(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Template) Delete(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Template) List(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *Template) Get(ctx *fiber.Ctx) error {
	panic("Implement me")
}
