package handler

import "github.com/gofiber/fiber/v2"

type Hello struct{}

func NewHello() *Hello {
	return &Hello{}
}

// @Summary      Hello
// @Tags         hello
// @Produce      plain
// @Success      200  {string}  string  "hello"
// @Security     BearerAuth
// @Router       /api/v1/internal/hello [get]
func (h *Hello) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("hello")
}
