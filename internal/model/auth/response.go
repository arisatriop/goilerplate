package auth

import (
	"goilerplate/internal/entity"

	"github.com/google/uuid"
)

type LoginResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	AccessToken         string    `json:"accessToken"`
	AccessTokenExpires  int       `json:"accessTokenExpiresIn"`
	RefreshToken        string    `json:"refreshToken"`
	RefreshTokenExpires int       `json:"refreshTokenExpiresIn"`
}

type TokenResponse struct {
	AccessToken          string `json:"accessToken"`
	AccessTokenExpiresAt int    `json:"expiresIn"`
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
