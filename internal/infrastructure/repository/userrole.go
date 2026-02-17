package repository

import (
	"context"
	"goilerplate/internal/domain/userrole"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type userroleRepo struct {
	db *gorm.DB
}

func NewUserRole(db *gorm.DB) userrole.Repository {
	return &userroleRepo{db: db}
}

func (r *userroleRepo) WithTx(ctx context.Context) userrole.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewUserRole(tx)
	}
	return r
}

func (r *userroleRepo) CreateUserRole(ctx context.Context, userRole *userrole.UserRole) error {
	now := utils.Now()

	// Get the current user ID from context for audit fields
	var createdBy string
	if val := ctx.Value(constants.ContextKeyUserID); val != nil {
		createdBy = val.(string)
	} else {
		createdBy = userRole.UserID.String() // Use the user ID being assigned the role
	}

	ur := &model.UserRole{
		UserID:    userRole.UserID,
		RoleID:    userRole.RoleID,
		CreatedAt: now,
		CreatedBy: createdBy,
	}

	if err := r.db.WithContext(ctx).Create(ur).Error; err != nil {
		return err
	}

	return nil
}
