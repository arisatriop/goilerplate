package middleware

import (
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	UserUsecase usecase.IUserUsecase
}

func NewAuth(userUsecase usecase.IUserUsecase) *Auth {
	return &Auth{
		UserUsecase: userUsecase,
	}
}

func (m *Auth) Authenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorization := ctx.Get("Authorization")
		if authorization == "" {
			return fiber.ErrUnauthorized
		}

		parts := strings.Split(authorization, " ")
		authorization = parts[0]
		if len(parts) == 2 {
			authorization = parts[1]
		}

		user, err := m.UserUsecase.GetByToken(ctx.UserContext(), authorization)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		if user == nil {
			return fiber.ErrUnauthorized
		}

		ctx.Locals("auth", &model.Auth{
			ID: user.ID,
		})

		return ctx.Next()
	}
}

func (m *Auth) Authorized(permission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// user := GetUser(ctx)

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden: '" + permission + "' permission is required",
		})
		// return ctx.Next()
	}
}

func GetUser(ctx *fiber.Ctx) *model.Auth {
	return ctx.Locals("auth").(*model.Auth)
}
