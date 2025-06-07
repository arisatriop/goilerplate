package repository

import (
	"context"
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model/auth"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IUserRepository interface {
	Create(ctx context.Context, db *gorm.DB, user *entity.User) error
	Update(ctx context.Context, db *gorm.DB, user *entity.User) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.User, error)
	GetByAccessToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error)
	GetByRefrehToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error)
	GetByEmail(ctx context.Context, db *gorm.DB, req *auth.LoginRequest) (*entity.User, error)
	GetByEmailAndPassword(ctx context.Context, db *gorm.DB, req *auth.LoginRequest) (*entity.User, error)
}

type UserRepository struct {
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) IUserRepository {
	return &UserRepository{
		Log: log,
	}
}

func (r *UserRepository) Create(ctx context.Context, db *gorm.DB, user *entity.User) error {
	if err := db.Create(user).Error; err != nil {
		r.Log.Errorf("failed to create user: %v", err)
		return err
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, db *gorm.DB, user *entity.User) error {
	if err := db.Save(user).Error; err != nil {
		r.Log.Errorf("failed to update user: %v", err)
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := db.Where("id = ? AND deleted_at is null", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByAccessToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("access_token = ? AND deleted_at is null", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by token: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByRefrehToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("refresh_token = ? AND deleted_at is null", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by refresh token: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, db *gorm.DB, req *auth.LoginRequest) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ? AND deleted_at is null", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by email and password: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmailAndPassword(ctx context.Context, db *gorm.DB, req *auth.LoginRequest) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ? AND password = ?", req.Email, req.Password).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by email and password: %v", err)
		return nil, err
	}
	return &user, nil
}
