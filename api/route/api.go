package route

import (
	"goilerplate/api/middleware"
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
)

func Init(fiberApp *fiber.App) {

	auth := fiberApp.Group("/auth")
	auth.Post("/token", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Token generated",
		})
	})
	auth.Post("/token/remove", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Session removed",
		})
	})

	app := config.GetAppVariable()
	api := fiberApp.Group("/api")
	middleware.Auth(api)

	v1 := api.Group("/v1")
	Example("/example", v1, app)

	v2 := api.Group("/v2")
	ExampleV2("/example", v2, app)

}
