package bootstrap

import (
	"goilerplate/config"
	"goilerplate/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewFiber(cfg *config.Config) *fiber.App {
	var app = fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: NewErrorHandler(),
		Prefork:      cfg.Server.Prefork,
		BodyLimit:    100 * 1024 * 1024, // 100MB limit for file uploads
	})

	app.Use(middleware.Recover())
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "*",
	// 	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	// 	AllowMethods: "*",
	// }))

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return ctx.Status(code).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
}
