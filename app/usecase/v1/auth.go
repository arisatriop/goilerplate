package v1

import (
	"goilerplate/app/entity"
	"goilerplate/config"
)

func NewAuthUsecase(app *config.App) IAuth {
	return &AuthImpl{
		App: app,
	}
}

type AuthImpl struct {
	App *config.App
}

type IAuth interface {
	Login() (*entity.Auth, error)
}

func (u *AuthImpl) Login() (*entity.Auth, error) {
	panic("not implemented") // TODO: Implement
}
