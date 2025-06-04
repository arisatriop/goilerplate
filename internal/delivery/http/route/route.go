package route

import (
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App  *fiber.App
	Auth *middleware.Auth

	ExampleController http.IExampleController
}

func (c *RouteConfig) Setup() {

	c.App.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to Goilerplate!")
	})
	c.App.Get("/health-check", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	api := c.App.Group("/api").Use(c.Auth.Authenticated())

	v1 := api.Group("/v1")
	v1.Post("/example", c.Auth.Authorized("example:create"), c.ExampleController.Create)
	v1.Get("/example", c.Auth.Authorized("example:list"), c.ExampleController.List)
	v1.Get("/example/:id", c.Auth.Authorized("example:view"), c.ExampleController.Get)
	v1.Put("/example/:id", c.Auth.Authorized("example:update"), c.ExampleController.Update)
	v1.Delete("/example/:id", c.Auth.Authorized("example:delete"), c.ExampleController.Delete)

}
