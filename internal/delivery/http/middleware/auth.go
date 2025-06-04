package middleware

import (
	"golang-clean-architecture/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (m *Auth) Authenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorization := ctx.Get("Authorization")
		if authorization == "" {
			return fiber.ErrUnauthorized
		}

		ctx.Locals("auth", &model.Auth{
			ID: uuid.New(),
		})
		return ctx.Next()
	}
}

func (m *Auth) Authorized(permission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		_ = GetUser(ctx)

		perm := "example"
		if permission[:7] != perm {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: '" + permission + "' permission is required",
			})
		}
		return ctx.Next()
	}
}

func GetUser(ctx *fiber.Ctx) *model.Auth {
	return ctx.Locals("auth").(*model.Auth)
}
