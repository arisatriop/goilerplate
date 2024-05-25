package route

import (
	"goilerplate/config"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func Init(app *fiber.App) {
	con := config.GetDBConnection()
	validator := validator.New()

	api := app.Group("/api")

	v1 := api.Group("/v1")
	Example("/example", v1, con, validator)

	v2 := api.Group("/v2")
	ExampleV2("/example", v2, con, validator)
}
