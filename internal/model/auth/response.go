package auth

import (
	"golang-clean-architecture/internal/entity"

	"github.com/google/uuid"
)

type LoginResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Token string    `json:"token"`
}

func ToLoginResponse(user *entity.User) *LoginResponse {
	return &LoginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Token: user.Token,
	}
}
