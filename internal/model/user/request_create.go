package user

import (
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/helper"
	"time"
)

type CreateRequest struct {
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	Password     string    `json:"password" validate:"required,min=8"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
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
