package handler

import "github.com/gofiber/fiber/v2"

type Hello struct{}

func NewHello() *Hello {
	return &Hello{}
}

// @Summary      Say hello
// @Tags         hello
// @Produce      plain
// @Success      200  {string}  string  "hello"
// @Failure      401  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/internal/hello [get]
func (h *Hello) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("hello")
}
