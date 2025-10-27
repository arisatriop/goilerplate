package router

import (
	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) index(ctx *fiber.Ctx) error {
	return ctx.SendString("Welcome to Goilerplate!")
}
