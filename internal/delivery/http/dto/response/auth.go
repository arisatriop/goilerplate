package dtoresponse

import "time"

// LoginResponse represents the login API response
type LoginResponse struct {
	User        UserResponse      `json:"user"`
	Permissions []string          `json:"permissions"`
	Tokens      TokenPairResponse `json:"tokens"`
	Session     SessionResponse   `json:"session"`
}

// UserResponse represents user data in API response
type UserResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Avatar        string     `json:"avatar"`
	IsActive      bool       `json:"isActive"`
	EmailVerified bool       `json:"emailVerified"`
	LastLoginAt   *time.Time `json:"lastLoginAt"`
}

// TokenPairResponse represents token data in API response
type TokenPairResponse struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenType       string    `json:"accessTokenType"`
	AccessTokenExpiresIn  int64     `json:"accessTokenExpiresIn"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenType      string    `json:"refreshTokenType"`
	RefreshTokenExpiresIn int64     `json:"refreshTokenExpiresIn"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

// SessionResponse represents session data in API response
type SessionResponse struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	DeviceType string    `json:"deviceType"`
	DeviceName string    `json:"deviceName"`
	IPAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	Location   string    `json:"location"`
	IsActive   bool      `json:"isActive"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}
