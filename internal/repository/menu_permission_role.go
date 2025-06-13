package repository

import (
	"context"
	"goilerplate/internal/entity"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MenuPermissionRoleRepository interface {
	BatchInsert(ctx context.Context, db *gorm.DB, mpr []entity.MenuPermissionRole) error
	DeleteByRoleID(ctx context.Context, db *gorm.DB, id uuid.UUID, deletedBy uuid.UUID) error
	HardDeleteByRoleID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type menuPermissionRoleRepository struct {
	Log *logrus.Logger
}

func NewMenuPermissionRoleRepository(log *logrus.Logger) MenuPermissionRoleRepository {
	return &menuPermissionRoleRepository{
		Log: log,
	}
}

func (r *menuPermissionRoleRepository) BatchInsert(ctx context.Context, db *gorm.DB, mpr []entity.MenuPermissionRole) error {
	if err := db.CreateInBatches(&mpr, 100).Error; err != nil {
		r.Log.Error("failed to create menu permission roles: ", err)
		return err
	}
	return nil
}

func (r *menuPermissionRoleRepository) DeleteByRoleID(ctx context.Context, db *gorm.DB, id uuid.UUID, deletedBy uuid.UUID) error {
	if err := db.Model(&entity.MenuPermissionRole{}).
		Where("role_id = ? and deleted_at is null", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error; err != nil {
		r.Log.Error("failed to delete menu permission roles by role ID: ", err)
		return err
	}
	return nil
}

func (r *menuPermissionRoleRepository) HardDeleteByRoleID(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	if err := db.Unscoped().Where("role_id = ?", id).Delete(&entity.MenuPermissionRole{}).Error; err != nil {
		r.Log.Error("failed to hard delete menu permission roles by role ID: ", err)
		return err
	}
	return nil
}
