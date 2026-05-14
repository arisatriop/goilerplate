package handler

import (
	pkgresponse "goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Hello struct{}

func NewHello() *Hello {
	return &Hello{}
}

func (h *Hello) Hello(ctx *fiber.Ctx) error {
	return pkgresponse.Success(ctx, "hello")
}
