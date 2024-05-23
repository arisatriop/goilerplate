package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Recovery(app *fiber.App) {

	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("panic: ", err)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    5001,
					"result":  false,
					"message": "Whops, something went wrong. Please try again in a moment",
					"data":    nil,
				})
			}
		}()
		return c.Next()
	})

}
