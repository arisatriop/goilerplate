package user

import (
	"goilerplate/internal/entity"
	"goilerplate/pkg/helper"
	"time"
)

type CreateRequest struct {
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	Password     string    `json:"password" validate:"required,min=8"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	CreatedAt    time.Time `json:"createdAt"`
	CreatedBy    string    `json:"createdBy"`
}

func (r *CreateRequest) ToCrete() *entity.User {
	return &entity.User{
		Name:         r.Name,
		Email:        r.Email,
		Password:     r.Password,
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		CreatedAt:    helper.NowJakarta(),
		CreatedBy:    "system",
	}
}
