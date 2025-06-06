package usecase

import (
	"context"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model/user"
	"golang-clean-architecture/internal/repository"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type IUserUsecase interface {
	Create(ctx context.Context, req *user.CreateRequest) error
	GetByToken(ctx context.Context, token string) (*entity.User, error)
}

type UserUsecase struct {
	Log            *logrus.Logger
	DB             *config.DB
	UserRepository repository.IUserRepository
}

func NewUserUsecase(log *logrus.Logger, db *config.DB, userRepo repository.IUserRepository) IUserUsecase {
	return &UserUsecase{
		Log:            log,
		DB:             db,
		UserRepository: userRepo,
	}
}

func (u *UserUsecase) Create(ctx context.Context, req *user.CreateRequest) error {

	passwordHashed, err := hasPassword(req.Password)
	if err != nil {
		u.Log.Errorf("failed to hash password: %v", err)
		return err
	}

	req.Password = passwordHashed
	user := req.ToCrete()

	if err := u.UserRepository.Create(ctx, u.DB.GDB, user); err != nil {
		u.Log.Errorf("failed to create user: %v", err)
		return err
	}

	return nil
}

func (u *UserUsecase) GetByToken(ctx context.Context, token string) (*entity.User, error) {
	user, err := u.UserRepository.GetByToken(ctx, u.DB.GDB, token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func hasPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}
