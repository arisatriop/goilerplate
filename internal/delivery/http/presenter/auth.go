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
		Menus:       ToMenusResponse(result.Menu),
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

// ToMenusResponse converts Menu entities to MenuResponse DTOs (recursive)
func ToMenusResponse(menus []auth.Menu) []dtoresponse.MenuResponse {
	result := make([]dtoresponse.MenuResponse, len(menus))
	for i, menu := range menus {
		result[i] = ToMenuResponse(menu)
	}
	return result
}

// ToMenuResponse converts a single Menu entity to MenuResponse DTO (recursive)
func ToMenuResponse(menu auth.Menu) dtoresponse.MenuResponse {
	return dtoresponse.MenuResponse{
		Name:         menu.Name,
		Slug:         menu.Slug,
		Icon:         menu.Icon,
		Route:        menu.Route,
		DisplayOrder: menu.DisplayOrder,
		IsActive:     menu.IsActive,
		Permissions:  menu.Permissions,
		Child:        ToMenusResponse(menu.Children), // Recursively map children
	}
}
