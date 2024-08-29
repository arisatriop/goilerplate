package request

import "github.com/gofiber/fiber/v2"

func NewAuthRequest() IAuth {
	return &AuthImpl{}
}

type AuthImpl struct{}

type IAuth interface {
	GetCreatePayload(ctx *fiber.Ctx) (*AuthCreatePayload, error)
}

type AuthCreatePayload struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

func (p *AuthImpl) GetCreatePayload(ctx *fiber.Ctx) (*AuthCreatePayload, error) {
	return &AuthCreatePayload{
		Username: ctx.FormValue("username"),
		Password: ctx.FormValue("password"),
	}, nil
}
