package repository

import (
	"context"
	"goilerplate/internal/model"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetPermission(ctx context.Context, db *gorm.DB, id uuid.UUID) (model.Permission, error)
}

type permissionRepository struct {
	Log *logrus.Logger
}

func NewPermissionRepository(log *logrus.Logger) PermissionRepository {
	return &permissionRepository{
		Log: log,
	}
}

func (r *permissionRepository) GetPermission(ctx context.Context, db *gorm.DB, id uuid.UUID) (model.Permission, error) {
	var permissions []string

	err := db.Table("users u").
		Select("mp.permission").
		Joins("JOIN role_users ru ON ru.user_id = u.id").
		Joins("JOIN menu_permission_roles mpr ON mpr.role_id = ru.role_id").
		Joins("JOIN menu_permissions mp ON mp.id = mpr.menu_permission_id").
		Joins("JOIN menus m ON m.id = mp.menu_id").
		Where("u.id = ?", id).
		Pluck("mp.permission", &permissions).Error

	if err != nil {
		r.Log.Errorf("failed to execute query: %v\n", err)
		return nil, err
	}

	permission := make(map[string]struct{})
	for _, name := range permissions {
		permission[name] = struct{}{}
	}

	return permission, nil
}
