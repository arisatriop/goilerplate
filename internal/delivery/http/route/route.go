package route

import (
	"golang-clean-architecture/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App  *fiber.App
	Auth *middleware.Auth
}

func (c *RouteConfig) Setup() {

	c.App.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to Goilerplate!")
	})

	api := c.App.Group("/api")
	api.Use(c.Auth.Authenticated())

	v1 := api.Group("/v1")
	v1.Get("/example", c.Auth.Authorized("example"), func(ctx *fiber.Ctx) error {
		return ctx.SendString("Example!")
	})
}
