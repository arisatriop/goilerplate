package repository

import (
	"context"
	"goilerplate/internal/entity"
	"goilerplate/internal/model/user"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository interface {
	Count(ctx context.Context, db *gorm.DB, req *user.GetRequest) (int64, error)
	Create(ctx context.Context, db *gorm.DB, user *entity.User) error
	Update(ctx context.Context, db *gorm.DB, user *entity.User) error
	GetAll(ctx context.Context, db *gorm.DB, req *user.GetRequest) ([]entity.User, int64, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.User, error)
	GetByAccessToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error)
	GetByRefrehToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error)
	GetByEmail(ctx context.Context, db *gorm.DB, email string) (*entity.User, error)
	GetByEmailAndPassword(ctx context.Context, db *gorm.DB, email, password string) (*entity.User, error)
}

type userRepository struct {
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) UserRepository {
	return &userRepository{
		Log: log,
	}
}

func (r *userRepository) Count(ctx context.Context, db *gorm.DB, req *user.GetRequest) (int64, error) {
	var count int64
	query := db.Model(&entity.User{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		keyword := "%" + strings.ToLower(req.Keyword) + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(name) LIKE ?", keyword, keyword)
	}

	if err := query.Count(&count).Error; err != nil {
		r.Log.Errorf("failed to count users: %v", err)
		return 0, err
	}
	return count, nil
}

func (r *userRepository) Create(ctx context.Context, db *gorm.DB, user *entity.User) error {
	if err := db.Create(user).Error; err != nil {
		r.Log.Errorf("failed to create user: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, db *gorm.DB, user *entity.User) error {
	if err := db.Save(user).Error; err != nil {
		r.Log.Errorf("failed to update user: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) GetAll(ctx context.Context, db *gorm.DB, req *user.GetRequest) ([]entity.User, int64, error) {
	var users []entity.User
	query := db.Model(&entity.User{}).Preload("Role").Where("deleted_at IS NULL")

	if req.Keyword != "" {
		keyword := "%" + strings.ToLower(req.Keyword) + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(name) LIKE ?", keyword, keyword)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if err := query.Find(&users).Error; err != nil {
		r.Log.Errorf("failed to get all users: %v", err)
		return nil, 0, err
	}

	total, err := r.Count(ctx, db, req)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := db.Preload("Role").Where("id = ? AND deleted_at is null", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByAccessToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error) {
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

func (r *userRepository) GetByRefrehToken(ctx context.Context, db *gorm.DB, token string) (*entity.User, error) {
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

func (r *userRepository) GetByEmail(ctx context.Context, db *gorm.DB, email string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ? AND deleted_at is null", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by email and password: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmailAndPassword(ctx context.Context, db *gorm.DB, email, password string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ? AND password = ?", email, password).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.Log.Errorf("failed to get user by email and password: %v", err)
		return nil, err
	}
	return &user, nil
}
