package pkg

import (
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
)

func NewFiber(cfg *config.Config) *fiber.App {
	var app = fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: NewErrorHandler(),
		Prefork:      cfg.Server.Prefork,
	})

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
