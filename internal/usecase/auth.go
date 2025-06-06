package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model/auth"
	"golang-clean-architecture/internal/repository"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type IAuthUsecase interface {
	Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, req *auth.LogoutRequest) error
}

type AuthUsecase struct {
	Log            *logrus.Logger
	DB             *config.DB
	UserRepository repository.IUserRepository
}

func NewAuthUsecase(log *logrus.Logger, db *config.DB, userRepo repository.IUserRepository) IAuthUsecase {
	return &AuthUsecase{
		Log:            log,
		DB:             db,
		UserRepository: userRepo,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	tx := u.DB.GDB.WithContext(ctx).Begin()

	user, err := u.UserRepository.GetByEmail(ctx, tx, req)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, helper.Error(http.StatusBadRequest, "invalid email or password")
	}

	if !isPasswordCorrect(req.Password, user.Password) {
		return nil, helper.Error(http.StatusBadRequest, "incorrect password")
	}

	user.Token = generateToken()
	if err := u.UserRepository.Update(ctx, tx, user); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit transaction:", err)
		return nil, err
	}

	return auth.ToLoginResponse(user), nil
}

func (u *AuthUsecase) Logout(ctx context.Context, req *auth.LogoutRequest) error {
	user, err := u.UserRepository.GetByID(ctx, u.DB.GDB, req.ID)
	if err != nil {
		return err
	}

	user.Token = ""
	if err := u.UserRepository.Update(ctx, u.DB.GDB, user); err != nil {
		return err
	}

	return nil
}

func isPasswordCorrect(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken() string {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return uuid.New().String() // Fallback to UUID if random generation fails
	}
	return hex.EncodeToString(bytes)
}
