package middleware

import (
	"errors"
	"fmt"
	"goilerplate/internal/helper"
	"goilerplate/internal/model"
	"goilerplate/internal/model/auth"
	"goilerplate/internal/usecase"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Auth struct {
	Config      *viper.Viper
	AuthUsecse  usecase.AuthUsecase
	UserUsecase usecase.UserUsecase
}

func NewAuth(config *viper.Viper, authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase) *Auth {
	return &Auth{
		Config:      config,
		AuthUsecse:  authUsecase,
		UserUsecase: userUsecase,
	}
}

func (m *Auth) Authenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorization := ctx.Get("Authorization")
		if authorization == "" {
			return model.Unauthorized(ctx)
		}

		parts := strings.Split(authorization, " ")
		token := parts[0]
		if len(parts) == 2 {
			token = parts[1]
		}

		jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.Config.GetString("jwt.secret")), nil
		})
		if err != nil || !jwtToken.Valid {
			return model.Unauthorized(ctx, "Unauthorized")
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok || claims["user_id"] == nil {
			return model.Unauthorized(ctx, "Invalid token claims")
		}

		user, err := m.UserUsecase.GetByAccessToken(ctx.UserContext(), token)
		if err != nil {
			return model.InternalServerError(ctx)
		}

		if user == nil {
			return model.Unauthorized(ctx)
		}
		ctx.Locals("auth", &auth.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})

		return ctx.Next()
	}
}

func (m *Auth) Authorized(permission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		key := fmt.Sprintf("permissions:%s", GetUser(ctx).ID)
		permissions, err := m.AuthUsecse.GetPermissionFromRedis(ctx.UserContext(), key)
		if err != nil {
			var cerr *helper.ClientError
			if errors.As(err, &cerr) {
				return model.JSON(ctx, cerr.Code, cerr.Message)
			}
			return model.InternalServerError(ctx)
		}

		if _, exists := permissions[permission]; exists {
			return ctx.Next()
		}

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden: '" + permission + "' permission is required",
		})
	}
}

func GetUser(ctx *fiber.Ctx) *auth.User {
	return ctx.Locals("auth").(*auth.User)
}
