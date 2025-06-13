package usecase

import (
	"context"
	"fmt"
	"goilerplate/internal/config"
	"goilerplate/internal/entity"
	"goilerplate/internal/helper"
	"goilerplate/internal/model/user"
	"goilerplate/internal/repository"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Get(ctx context.Context, id uuid.UUID) (*user.GetResponse, error)
	GetAll(ctx context.Context, req *user.GetRequest) ([]user.GetAllResponse, int64, error)
	Create(ctx context.Context, req *user.CreateRequest) error
	Update(ctx context.Context, req *user.UpdateRequest) error
	Delete(ctx context.Context, req *user.DeleteRequest) error
	GetByAccessToken(ctx context.Context, token string) (*entity.User, error)
}

type userUsecase struct {
	Log            *logrus.Logger
	DB             *config.DB
	UserRepository repository.UserRepository
}

func NewUserUsecase(log *logrus.Logger, db *config.DB, userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{
		Log:            log,
		DB:             db,
		UserRepository: userRepo,
	}
}

func (u *userUsecase) GetAll(ctx context.Context, req *user.GetRequest) ([]user.GetAllResponse, int64, error) {
	users, total, err := u.UserRepository.GetAll(ctx, u.DB.GDB.WithContext(ctx), req)
	if err != nil {
		return nil, 0, err
	}

	var response []user.GetAllResponse
	for _, u := range users {
		response = append(response, *user.ToGetAll(&u))
	}

	return response, total, nil
}

func (u *userUsecase) Get(ctx context.Context, id uuid.UUID) (*user.GetResponse, error) {
	usr, err := u.UserRepository.GetByID(ctx, u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return nil, err
	}

	if usr == nil {
		return nil, helper.Error(http.StatusNotFound, fmt.Sprintf("user with ID %s not found", id))
	}

	return user.ToGet(usr), nil
}

func (u *userUsecase) Create(ctx context.Context, req *user.CreateRequest) error {

	passwordHashed, err := hasPassword(req.Password)
	if err != nil {
		u.Log.Errorf("failed to hash password: %v", err)
		return err
	}

	req.Password = passwordHashed
	user := req.ToCrete()

	return u.UserRepository.Create(ctx, u.DB.GDB.WithContext(ctx), user)
}

func (u *userUsecase) Update(ctx context.Context, req *user.UpdateRequest) error {
	usr, err := u.UserRepository.GetByID(ctx, u.DB.GDB.WithContext(ctx), req.ID)
	if err != nil {
		return err
	}
	if usr == nil {
		return helper.Error(http.StatusNotFound, fmt.Sprintf("user with ID %s not found", req.ID))
	}

	usr, err = u.UserRepository.GetByEmail(ctx, u.DB.GDB.WithContext(ctx), req.Email)
	if err != nil {
		return err
	}
	if usr != nil && usr.ID != req.ID {
		return helper.Error(http.StatusConflict, fmt.Sprintf("email '%s' is already in use", req.Email))
	}

	passwordHashed, err := hasPassword(req.Password)
	if err != nil {
		u.Log.Errorf("failed to hash password: %v", err)
		return err
	}

	req.Password = passwordHashed
	req.ToUpdate(usr)

	return u.UserRepository.Update(ctx, u.DB.GDB.WithContext(ctx), usr)
}

func (u *userUsecase) Delete(ctx context.Context, req *user.DeleteRequest) error {
	usr, err := u.UserRepository.GetByID(ctx, u.DB.GDB.WithContext(ctx), req.ID)
	if err != nil {
		return err
	}
	if usr == nil {
		return helper.Error(http.StatusNotFound, fmt.Sprintf("user with ID %s not found", req.ID))
	}

	if usr.ID == req.DeletedBy {
		return helper.Error(http.StatusBadRequest, "you cannot delete yourself")
	}

	req.ToDelete(usr)
	return u.UserRepository.Update(ctx, u.DB.GDB.WithContext(ctx), usr)
}

func (u *userUsecase) GetByAccessToken(ctx context.Context, token string) (*entity.User, error) {
	user, err := u.UserRepository.GetByAccessToken(ctx, u.DB.GDB.WithContext(ctx), token)
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
