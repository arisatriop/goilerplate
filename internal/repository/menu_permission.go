package repository

import (
	"context"
	"goilerplate/internal/entity"
	"goilerplate/internal/model/menupermission"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MenuPermissionRepository interface {
	GetAll(ctx context.Context, db *gorm.DB, req *menupermission.GetRequest) ([]entity.MenuPermission, int64, error)
	Count(ctx context.Context, db *gorm.DB, req *menupermission.GetRequest) (int64, error)
}

type menuPermissionRepository struct {
	Log *logrus.Logger
}

func NewMenuPermissionRepository(log *logrus.Logger) MenuPermissionRepository {
	return &menuPermissionRepository{
		Log: log,
	}
}

func (r *menuPermissionRepository) GetAll(ctx context.Context, db *gorm.DB, req *menupermission.GetRequest) ([]entity.MenuPermission, int64, error) {
	query := db.Model(&entity.MenuPermission{}).
		Preload("Menu").
		Joins("JOIN menus ON menus.id = menu_permissions.menu_id").
		Where("menu_permissions.deleted_at IS NULL")

	if req.Keyword != "" {
		query = query.Where("LOWER(menus.name) LIKE ?", "%"+strings.ToLower(req.Keyword)+"%")
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	var menuPermissions []entity.MenuPermission
	if err := query.Find(&menuPermissions).Error; err != nil {
		r.Log.Error("Failed to get menu permissions: ", err)
		return nil, 0, err
	}

	count, err := r.Count(ctx, db, req)
	if err != nil {
		return nil, 0, err
	}

	return menuPermissions, count, nil
}

func (r *menuPermissionRepository) Count(ctx context.Context, db *gorm.DB, req *menupermission.GetRequest) (int64, error) {
	query := db.Model(&entity.MenuPermission{}).
		Joins("JOIN menus ON menus.id = menu_permissions.menu_id").
		Where("menu_permissions.deleted_at IS NULL")

	if req.Keyword != "" {
		query = query.Where("LOWER(menus.name) LIKE ?", "%"+strings.ToLower(req.Keyword)+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.Log.Error("Failed to count menu permissions: ", err)
		return 0, err
	}

	return count, nil
}
