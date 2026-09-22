// This file is a TEMPLATE, not a feature.
//
// foo is the blank scaffold a new domain is copied from; bar next to it is the worked example
// with the bodies filled in. The panics are deliberate — they are what an unimplemented method
// should do in a template, because a silent zero value would let a half-copied domain look like
// it works. Its HTTP routes are not registered; see internal/delivery/http/router/public.go.
//
// The copy procedure is in .claude/skills/crud-operations/SKILL.md.

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
