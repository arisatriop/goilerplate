package repository

import (
	"context"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUser(db *gorm.DB) user.Repository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) WithTx(ctx context.Context) user.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewUser(tx)
	}
	return r
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).
		Select("id", "name", "phone", "email", "avatar", "is_active", "password_hash").
		Where("email = ?", email).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomainEntity(&u), nil
}

func (r *userRepo) CreateUser(ctx context.Context, usr *user.User) (*user.User, error) {
	now := utils.Now()
	userID := utils.GenerateUUID()

	var createdBy string
	if val := ctx.Value(constants.ContextKeyUserID); val != nil {
		createdBy = val.(string)
	} else {
		createdBy = userID // Use the new user ID if no context user
	}

	u := &model.User{
		ID:                userID,
		Name:              usr.Name,
		Phone:             usr.Phone,
		Email:             usr.Email,
		Avatar:            usr.Avatar,
		IsActive:          usr.IsActive,
		PasswordHash:      usr.PasswordHash,
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         createdBy,
		UpdatedBy:         createdBy,
		PasswordChangedAt: now,
	}

	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntity(u), nil
}

func (r *userRepo) toDomainEntity(u *model.User) *user.User {
	id, _ := uuid.Parse(u.ID)
	return &user.User{
		ID:           id,
		Name:         u.Name,
		Phone:        u.Phone,
		Email:        u.Email,
		Avatar:       u.Avatar,
		IsActive:     u.IsActive,
		PasswordHash: u.PasswordHash,
	}
}
