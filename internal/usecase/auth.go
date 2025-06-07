package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/model/auth"
	"golang-clean-architecture/internal/repository"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type IAuthUsecase interface {
	Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error)
	Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, req *auth.LogoutRequest) error
	SetPermission(ctx context.Context, id uuid.UUID) error
	GetPermissions(ctx context.Context, id uuid.UUID) (model.Permission, error)
	GetPermissionFromRedis(ctx context.Context, key string) (map[string]struct{}, error)
}

type AuthUsecase struct {
	Config               *viper.Viper
	Log                  *logrus.Logger
	DB                   *config.DB
	UserRepository       repository.IUserRepository
	PermissionRepository repository.IPermissionRepository
	RedisRepository      *repository.RedisRepository
}

func NewAuthUsecase(
	viper *viper.Viper,
	log *logrus.Logger,
	db *config.DB,
	userRepo repository.IUserRepository,
	permissionRepo repository.IPermissionRepository,
	redisRepo *repository.RedisRepository) IAuthUsecase {
	return &AuthUsecase{
		Log:                  log,
		Config:               viper,
		DB:                   db,
		UserRepository:       userRepo,
		PermissionRepository: permissionRepo,
		RedisRepository:      redisRepo,
	}
}

func (u *AuthUsecase) Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	user, err := u.UserRepository.GetByRefrehToken(ctx, u.DB.GDB, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, helper.Error(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (any, error) {
		return []byte(u.Config.GetString("jwt.refresh_secret")), nil
	})
	if err != nil || !token.Valid {
		u.Log.Error("failed to parse refresh token: ", err)
		return nil, helper.Error(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil || claims["token_id"] == nil {
		return nil, helper.Error(http.StatusUnauthorized, "invalid token claims")
	}

	userID := uuid.MustParse(claims["user_id"].(string))
	accessToken, err := GenerateAccessToken(userID, viper.GetString("jwt.secret"), time.Duration(viper.GetInt("jwt.access_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	user.AccessToken = accessToken
	if err := u.UserRepository.Update(ctx, u.DB.GDB, user); err != nil {
		return nil, err
	}

	return auth.ToTokenResponse(accessToken, u.Config.GetInt("jwt.access_token_expiry")), nil
}

func (u *AuthUsecase) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {

	user, err := u.UserRepository.GetByEmail(ctx, u.DB.GDB, req)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, helper.Error(http.StatusBadRequest, "invalid email or password")
	}

	if !isPasswordCorrect(req.Password, user.Password) {
		return nil, helper.Error(http.StatusBadRequest, "incorrect password")
	}

	accessToken, err := GenerateAccessToken(
		user.ID,
		u.Config.GetString("jwt.secret"),
		time.Duration(u.Config.GetInt("jwt.access_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := GenerateRefreshToken(
		user.ID,
		u.Config.GetString("jwt.refresh_secret"),
		time.Duration(u.Config.GetInt("jwt.refresh_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// key := fmt.Sprintf("refresh:%s:%s", user.ID, tokenID)
	// err = u.DB.Redis.Set(ctx, key, "valid", time.Duration(viper.GetInt("jwt.refresh_token_expiry"))*time.Second).Err()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to set refresh token in Redis: %w", err)
	// }

	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	if err := u.UserRepository.Update(ctx, u.DB.GDB, user); err != nil {
		return nil, err
	}

	if err := u.SetPermission(ctx, user.ID); err != nil {
		return nil, err
	}

	return auth.ToLoginResponse(
		user,
		u.Config.GetInt("jwt.access_token_expiry"),
		u.Config.GetInt("jwt.refresh_token_expiry"),
	), nil
}

func (u *AuthUsecase) Logout(ctx context.Context, req *auth.LogoutRequest) error {
	user, err := u.UserRepository.GetByID(ctx, u.DB.GDB, req.ID)
	if err != nil {
		return err
	}

	user.AccessToken = ""
	user.RefreshToken = ""
	if err := u.UserRepository.Update(ctx, u.DB.GDB, user); err != nil {
		return err
	}

	return nil
}

func (u *AuthUsecase) SetPermission(ctx context.Context, id uuid.UUID) error {
	permissions, err := u.PermissionRepository.GetPermission(ctx, u.DB.PgxPool, id)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("permissions:%s", id.String())
	value, _ := json.Marshal(permissions)
	if err := u.RedisRepository.Set(ctx, key, value, time.Duration(u.Config.GetInt("jwt.access_token_expiry"))*time.Second); err != nil {
		u.Log.Error("failed to set permissions in Redis: ", err)
		return err
	}

	return nil
}

func (u *AuthUsecase) GetPermissions(ctx context.Context, id uuid.UUID) (model.Permission, error) {
	return u.PermissionRepository.GetPermission(ctx, u.DB.PgxPool, id)
}

func (u *AuthUsecase) GetPermissionFromRedis(ctx context.Context, key string) (map[string]struct{}, error) {
	permission, err := u.RedisRepository.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, helper.Error(http.StatusUnauthorized, "Unauthorized")
		}
		u.Log.Errorf("failed to get permission from redis: %v\n", err)
		return nil, err
	}

	var permissions map[string]struct{}
	if err := json.Unmarshal([]byte(permission), &permissions); err != nil {
		u.Log.Errorf("failed to unmarshal permissions: %v\n", err)
		return nil, err
	}

	return permissions, nil
}

func GenerateAccessToken(userID uuid.UUID, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uuid.UUID, secret string, duration time.Duration) (string, string, error) {
	tokenID := uuid.NewString()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"token_id": tokenID,
		"exp":      time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	return signedToken, tokenID, err
}

func isPasswordCorrect(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
