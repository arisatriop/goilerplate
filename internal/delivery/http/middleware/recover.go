package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// You can log the panic if needed
				fmt.Printf("Recovered from panic: %v\n", r)

				// Respond with internal server error
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Whoops! Something went wrong.",
				})
			}
		}()

		// Continue to next handler
		return c.Next()
	}
}
