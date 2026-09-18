package jwt

import (
	"errors"
	"fmt"
	"time"

	"goilerplate/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = utils.ClientErr(401, "Invalid token")
	ErrExpiredToken = utils.ClientErr(401, "Token has expired")
	ErrTokenClaims  = utils.ClientErr(401, "Invalid token claims")
)

// Token types, carried in the "type" claim.
const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

// SigningMethod is the only algorithm accepted when verifying. Tokens signed with anything
// else — including "none" — are rejected before their signature is even considered.
var SigningMethod = jwt.SigningMethodHS256

// DefaultLeeway absorbs small clock differences between instances when no leeway is configured.
const DefaultLeeway = 30 * time.Second

// Key is one signing key pair, identified by ID. The ID is published in each token's "kid"
// header so a token can still be verified after the active key has moved on.
type Key struct {
	ID            string
	AccessSecret  string
	RefreshSecret string
}

// Config describes how tokens are signed and verified.
type Config struct {
	// Active signs every newly issued token.
	Active Key
	// Previous keys are accepted for verification only, so secrets can be rotated without
	// invalidating tokens that are still within their lifetime.
	Previous []Key
	Issuer   string
	Audience string
	// AccessExpiry is the lifetime of an access token. The refresh token lifetime is passed
	// per call instead, because it belongs to the session (remember me changes it).
	AccessExpiry time.Duration
	Leeway       time.Duration
}

type JWTService struct {
	active       Key
	keys         map[string]Key
	issuer       string
	audience     string
	accessExpiry time.Duration
	leeway       time.Duration
}

type TokenPair struct {
	AccessToken           string
	AccessTokenID         string // jti of the access token
	AccessTokenType       string
	AccessTokenExpiresIn  int64
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenID        string // jti of the refresh token; stored as user_sessions.refresh_jti
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

// NewJWTService builds the service from cfg. It fails when the active key is unusable or
// when a previous key reuses the active key's ID, which would make "kid" ambiguous.
func NewJWTService(cfg Config) (*JWTService, error) {
	if cfg.Active.ID == "" {
		return nil, errors.New("jwt: active key id is required")
	}
	if cfg.Active.AccessSecret == "" || cfg.Active.RefreshSecret == "" {
		return nil, errors.New("jwt: active key needs both an access and a refresh secret")
	}
	if cfg.Issuer == "" {
		return nil, errors.New("jwt: issuer is required")
	}
	if cfg.Audience == "" {
		return nil, errors.New("jwt: audience is required")
	}
	if cfg.AccessExpiry <= 0 {
		return nil, errors.New("jwt: access expiry must be greater than 0")
	}

	keys := map[string]Key{cfg.Active.ID: cfg.Active}
	for _, key := range cfg.Previous {
		if key.ID == "" {
			return nil, errors.New("jwt: every previous key needs an id")
		}
		if _, exists := keys[key.ID]; exists {
			return nil, fmt.Errorf("jwt: duplicate key id %q", key.ID)
		}
		keys[key.ID] = key
	}

	leeway := cfg.Leeway
	if leeway <= 0 {
		leeway = DefaultLeeway
	}

	return &JWTService{
		active:       cfg.Active,
		keys:         keys,
		issuer:       cfg.Issuer,
		audience:     cfg.Audience,
		accessExpiry: cfg.AccessExpiry,
		leeway:       leeway,
	}, nil
}

// GenerateTokenPair creates an access and a refresh token for a new session.
// refreshExpiry is the session's absolute lifetime, which the caller owns.
func (j *JWTService) GenerateTokenPair(
	userID, userName, email, sessionID, deviceID string,
	refreshExpiry time.Duration,
) (*TokenPair, error) {
	now := utils.Now()

	accessID := utils.GenerateUUID()
	accessExpiresAt := now.Add(j.accessExpiry)
	accessToken, err := j.sign(&Claims{
		UserID:           userID,
		UserName:         userName,
		Email:            email,
		SessionID:        sessionID,
		DeviceID:         deviceID,
		Type:             AccessToken,
		RegisteredClaims: j.registered(userID, accessID, now, accessExpiresAt),
	}, j.active.AccessSecret)
	if err != nil {
		return nil, err
	}

	// The refresh token carries minimal claims: it is only ever exchanged, never read for
	// identity. Its jti is what user_sessions tracks and rotates.
	refreshID := utils.GenerateUUID()
	refreshExpiresAt := now.Add(refreshExpiry)
	refreshToken, err := j.sign(&Claims{
		UserID:           userID,
		SessionID:        sessionID,
		DeviceID:         deviceID,
		Type:             RefreshToken,
		RegisteredClaims: j.registered(userID, refreshID, now, refreshExpiresAt),
	}, j.active.RefreshSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           accessToken,
		AccessTokenID:         accessID,
		AccessTokenType:       "Bearer",
		AccessTokenExpiresIn:  int64(j.accessExpiry.Seconds()),
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenID:        refreshID,
		RefreshTokenType:      "Bearer",
		RefreshTokenExpiresIn: int64(refreshExpiry.Seconds()),
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// GenerateAccessToken creates only an access token, for the refresh flow.
func (j *JWTService) GenerateAccessToken(userID, userName, email, sessionID, deviceID string) (string, time.Time, error) {
	now := utils.Now()
	expiresAt := now.Add(j.accessExpiry)

	token, err := j.sign(&Claims{
		UserID:           userID,
		UserName:         userName,
		Email:            email,
		SessionID:        sessionID,
		DeviceID:         deviceID,
		Type:             AccessToken,
		RegisteredClaims: j.registered(userID, utils.GenerateUUID(), now, expiresAt),
	}, j.active.AccessSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

// ValidateAccessToken verifies an access token and rejects a refresh token presented in its place.
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validate(tokenString, AccessToken)
}

// ValidateRefreshToken verifies a refresh token and rejects an access token presented in its place.
func (j *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validate(tokenString, RefreshToken)
}

// AccessTokenExpiry returns the configured access token lifetime.
func (j *JWTService) AccessTokenExpiry() time.Duration {
	return j.accessExpiry
}

func (j *JWTService) registered(userID, tokenID string, now, expiresAt time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    j.issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{j.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(now),
		ID:        tokenID,
	}
}

func (j *JWTService) sign(claims *Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(SigningMethod, claims)
	token.Header["kid"] = j.active.ID
	return token.SignedString([]byte(secret))
}

// validate parses the token with the secret named by its "kid" header. Only the secret of
// the requested token type is offered, so a refresh token can never satisfy an access-token
// check even if both were signed by the same key.
func (j *JWTService) validate(tokenString, wantType string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrInvalidToken
		}

		key, known := j.keys[kid]
		if !known {
			return nil, ErrInvalidToken
		}

		if wantType == RefreshToken {
			return []byte(key.RefreshSecret), nil
		}
		return []byte(key.AccessSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc,
		jwt.WithValidMethods([]string{SigningMethod.Alg()}),
		jwt.WithIssuer(j.issuer),
		jwt.WithAudience(j.audience),
		jwt.WithLeeway(j.leeway),
		jwt.WithExpirationRequired(),
	)
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

	if claims.Type != wantType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
