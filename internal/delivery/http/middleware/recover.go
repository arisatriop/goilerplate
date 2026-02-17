package middleware

import (
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Set panic flag for request logger
				c.Locals("panic_occurred", true)
				c.Locals("panic_value", r)

				// Get request ID if available
				requestID := ""
				if id := c.Locals(string(constants.ContextKeyRequestID)); id != nil {
					requestID = id.(string)
				}

				// Log the panic with structured logging
				slog.ErrorContext(c.Context(),
					"Panic recovered in HTTP handler",
					slog.String("request_id", requestID),
					slog.String("method", c.Method()),
					slog.String("path", c.Path()),
					slog.String("url", c.OriginalURL()),
					slog.Any("panic_value", r),
					slog.String("stack_trace", string(debug.Stack())),
				)

				// Respond with standardized error response
				response.InternalServerError(c, "An unexpected error occurred")
			}
		}()

		// Continue to next handler
		return c.Next()
	}
}
