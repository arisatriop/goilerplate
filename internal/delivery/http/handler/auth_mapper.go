package handler

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/jwt"
)

// AuthResponseMapper handles mapping between domain entities and response DTOs
type AuthResponseMapper struct{}

// NewAuthResponseMapper creates a new response mapper
func NewAuthResponseMapper() *AuthResponseMapper {
	return &AuthResponseMapper{}
}

// MapLoginResult converts LoginResult to LoginResponse DTO
func (m *AuthResponseMapper) MapLoginResult(result *auth.LoginResult) *dtoresponse.LoginResponse {
	return &dtoresponse.LoginResponse{
		User:       m.mapUser(result.User),
		Menu:       m.mapMenus(result.Menus),
		CustomMenu: m.mapMenus(result.CustomMenus),
		Tokens:     m.mapTokenPair(result.Tokens),
		Session:    m.mapSession(result.Session),
	}
}

// mapUser converts User entity to UserResponse DTO
func (m *AuthResponseMapper) mapUser(user *auth.User) dtoresponse.UserResponse {
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

// mapTokenPair converts JWT TokenPair to TokenPairResponse DTO
func (m *AuthResponseMapper) mapTokenPair(tokens *jwt.TokenPair) dtoresponse.TokenPairResponse {
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

// mapSession converts UserSession entity to SessionResponse DTO
func (m *AuthResponseMapper) mapSession(session *auth.UserSession) dtoresponse.SessionResponse {
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

// mapMenus converts Menu entities to MenuResponse DTOs (recursive)
func (m *AuthResponseMapper) mapMenus(menus []auth.Menu) []dtoresponse.MenuResponse {
	result := make([]dtoresponse.MenuResponse, len(menus))
	for i, menu := range menus {
		result[i] = m.mapMenu(menu)
	}
	return result
}

// mapMenu converts a single Menu entity to MenuResponse DTO (recursive)
func (m *AuthResponseMapper) mapMenu(menu auth.Menu) dtoresponse.MenuResponse {
	return dtoresponse.MenuResponse{
		Name:         menu.Name,
		Slug:         menu.Slug,
		Icon:         menu.Icon,
		Route:        menu.Route,
		DisplayOrder: menu.DisplayOrder,
		IsActive:     menu.IsActive,
		Child:        m.mapMenus(menu.Children), // Recursively map children
	}
}
