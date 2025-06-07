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
	AuthController    http.IAuthController
	UserController    http.IUserController
}

func (c *RouteConfig) Setup() {

	c.App.Use(middleware.Recover())

	c.App.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to Goilerplate!")
	})
	c.App.Get("/health-check", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.Post("/token", c.AuthController.Token)
	auth.Post("/login", c.AuthController.Login)
	auth.Post("/logout", c.Auth.Authenticated(), c.AuthController.Logout)

	examples := v1.Group("/examples").Use(c.Auth.Authenticated())
	examples.Post("/", c.Auth.Authorized("example:create"), c.ExampleController.Create)
	examples.Get("/", c.Auth.Authorized("example:list"), c.ExampleController.GetAll)
	examples.Get("/:id", c.Auth.Authorized("example:view"), c.ExampleController.Get)
	examples.Put("/:id", c.Auth.Authorized("example:update"), c.ExampleController.Update)
	examples.Delete("/:id", c.Auth.Authorized("example:delete"), c.ExampleController.Delete)

	users := v1.Group("/users").Use(c.Auth.Authenticated())
	users.Post("/", c.Auth.Authorized("manage.user:create"), c.UserController.Create)

}
