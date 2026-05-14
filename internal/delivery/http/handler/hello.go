package handler

import "github.com/gofiber/fiber/v2"

type Hello struct{}

func NewHello() *Hello {
	return &Hello{}
}

func (h *Hello) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("hello")
}
