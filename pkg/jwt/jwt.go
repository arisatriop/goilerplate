package jwt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"goilerplate/pkg/utils"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = utils.Error(401, "Invalid token")
	ErrExpiredToken = utils.Error(401, "Token has expired")
	ErrTokenClaims  = utils.Error(401, "Invalid token claims")
)

// Token types
const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

type JWTService struct {
	secretKey     string // General secret key (fallback)
	accessSecret  string // Secret for access tokens
	refreshSecret string // Secret for refresh tokens
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

type TokenPair struct {
	AccessToken           string
	AccessTokenType       string
	AccessTokenExpiresIn  int64
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenType      string
	RefreshTokenExpiresIn int64
	RefreshTokenExpiresAt time.Time
}

type Claims struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	DeviceID  string `json:"device_id,omitempty"`
	Type      string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func NewJWTService(secretKey, accessSecret, refreshSecret, issuer string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

// GenerateTokenPair creates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID, userName, email, sessionID, deviceID string) (*TokenPair, error) {
	now := utils.Now()

	// Generate Access Token
	accessClaims := &Claims{
		UserID:    userID,
		UserName:  userName,
		Email:     email,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Type:      AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecret))
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token (longer expiry, minimal claims)
	refreshClaims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Type:      RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.refreshSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           accessTokenString,
		AccessTokenType:       "Bearer",
		AccessTokenExpiresIn:  int64(j.accessExpiry.Seconds()),
		AccessTokenExpiresAt:  now.Add(j.accessExpiry),
		RefreshToken:          refreshTokenString,
		RefreshTokenType:      "Bearer",
		RefreshTokenExpiresIn: int64(j.refreshExpiry.Seconds()),
		RefreshTokenExpiresAt: now.Add(j.refreshExpiry),
	}, nil
}

// GenerateAccessToken creates only an access token (for refresh scenarios)
func (j *JWTService) GenerateAccessToken(userID, userName, email, sessionID, deviceID string) (string, time.Time, error) {
	now := utils.Now()
	expiresAt := now.Add(j.accessExpiry)

	// Generate Access Token with unique JTI to prevent duplicates
	accessClaims := &Claims{
		UserID:    userID,
		UserName:  userName,
		Email:     email,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Type:      AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateUniqueID(), // Add unique JTI to prevent identical tokens
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return accessTokenString, expiresAt, nil
}

// ValidateToken validates and parses a JWT token (auto-detects token type)
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// First try to extract claims to determine token type
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Validate with appropriate secret based on token type
	switch claims.Type {
	case AccessToken:
		return j.ValidateAccessToken(tokenString)
	case RefreshToken:
		return j.ValidateRefreshToken(tokenString)
	default:
		return j.validateTokenWithSecret(tokenString, j.secretKey)
	}
}

// ValidateAccessToken validates an access token specifically
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validateTokenWithSecret(tokenString, j.accessSecret)
}

// ValidateRefreshToken validates a refresh token specifically
func (j *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validateTokenWithSecret(tokenString, j.refreshSecret)
}

// validateTokenWithSecret is a helper method to validate tokens with a specific secret
func (j *JWTService) validateTokenWithSecret(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenClaims
	}

	return claims, nil
}

// ExtractClaims extracts claims without validation (useful for expired tokens)
func (j *JWTService) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenClaims
	}

	return claims, nil
}

// generateUniqueID generates a unique identifier for JWT tokens
func generateUniqueID() string {
	// Use timestamp in nanoseconds + cryptographically secure random number for uniqueness
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(9223372036854775807)) // max int64
	return fmt.Sprintf("%d-%s", utils.Now().UnixNano(), randomNum.String())
}
