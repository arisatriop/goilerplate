package route

import (
	"goilerplate/config"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// type RouteGroup struct {
// 	Public *gin.RouterGroup
// }

// func InitRoute(e *gin.Engine) {
// 	con := config.GetDBConnection()
// 	validator := validator.New()

// 	Example(e, con, validator)
// }

func Init(app *fiber.App) {
	con := config.GetDBConnection()
	validator := validator.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	Example("/example", v1, con, validator)
}
