package handler

import (
	"goilerplate/internal/domain/zexamplenew"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ZexampleNew struct {
	Validator *validator.Validate
	Usecase   zexamplenew.Usecase
}

func NewZexampleNew(validator *validator.Validate, usecase zexamplenew.Usecase) *ZexampleNew {
	return &ZexampleNew{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *ZexampleNew) Create(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *ZexampleNew) Update(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *ZexampleNew) Delete(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *ZexampleNew) List(ctx *fiber.Ctx) error {
	panic("Implement me")
}

func (h *ZexampleNew) Get(ctx *fiber.Ctx) error {
	panic("Implement me")
}
