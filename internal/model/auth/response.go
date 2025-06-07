package auth

import (
	"golang-clean-architecture/internal/entity"

	"github.com/google/uuid"
)

type LoginResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	AccessToken         string    `json:"access_token"`
	AccessTokenExpires  int       `json:"access_token_expires_in"`
	RefreshToken        string    `json:"refresh_token"`
	RefreshTokenExpires int       `json:"refresh_token_expires_in"`
}

type TokenResponse struct {
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresAt int    `json:"expires_in"`
}

func ToLoginResponse(user *entity.User, xAcess, xRefresh int) *LoginResponse {
	return &LoginResponse{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		AccessToken:         user.AccessToken,
		AccessTokenExpires:  xAcess,
		RefreshToken:        user.RefreshToken,
		RefreshTokenExpires: xRefresh,
	}
}

func ToTokenResponse(token string, expires int) *TokenResponse {
	return &TokenResponse{
		AccessToken:          token,
		AccessTokenExpiresAt: expires,
	}
}
