package route

import (
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
)

func Init(fiberApp *fiber.App) {
	app := config.GetAppVariable()

	api := fiberApp.Group("/api")

	v1 := api.Group("/v1")
	Example("/example", v1, app)

	v2 := api.Group("/v2")
	ExampleV2("/example", v2, app)
}
