package handler

import (
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Hello struct{}

func NewHello() *Hello {
	return &Hello{}
}

// @Summary      Hello
// @Tags         hello
// @Produce      json
// @Success      200  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/internal/hello [get]
func (h *Hello) Hello(ctx *fiber.Ctx) error {
	return response.Success(ctx, nil, response.WithMessage("hello"))
}
