package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/jwt"
)

// ToLoginResponse converts LoginResult to LoginResponse DTO
func ToLoginResponse(result *auth.LoginResult) *dtoresponse.LoginResponse {
	return &dtoresponse.LoginResponse{
		User:        ToUserResponse(result.User),
		Permissions: result.Permission,
		Tokens:      ToTokenPairResponse(result.Tokens),
		Session:     ToSessionResponse(result.Session),
	}
}

// ToUserResponse converts User entity to UserResponse DTO
func ToUserResponse(user *auth.User) dtoresponse.UserResponse {
	return dtoresponse.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Avatar:        user.Avatar,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
	}
}

// ToTokenPairResponse converts JWT TokenPair to TokenPairResponse DTO
func ToTokenPairResponse(tokens *jwt.TokenPair) dtoresponse.TokenPairResponse {
	return dtoresponse.TokenPairResponse{
		AccessToken:           tokens.AccessToken,
		AccessTokenType:       tokens.AccessTokenType,
		AccessTokenExpiresIn:  tokens.AccessTokenExpiresIn,
		AccessTokenExpiresAt:  tokens.AccessTokenExpiresAt,
		RefreshToken:          tokens.RefreshToken,
		RefreshTokenType:      tokens.RefreshTokenType,
		RefreshTokenExpiresIn: tokens.RefreshTokenExpiresIn,
		RefreshTokenExpiresAt: tokens.RefreshTokenExpiresAt,
	}
}

// ToSessionResponse converts UserSession entity to SessionResponse DTO
func ToSessionResponse(session *auth.UserSession) dtoresponse.SessionResponse {
	return dtoresponse.SessionResponse{
		ID:         session.ID,
		DeviceID:   session.DeviceID,
		DeviceType: session.DeviceType,
		DeviceName: session.DeviceName,
		IPAddress:  session.IPAddress,
		UserAgent:  session.UserAgent,
		Location:   session.Location,
		IsActive:   session.IsActive,
		ExpiresAt:  session.ExpiresAt,
		LastUsedAt: session.LastUsedAt,
	}
}
