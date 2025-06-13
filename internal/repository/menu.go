package repository

import (
	"context"
	"goilerplate/internal/entity"
	"goilerplate/internal/model/menu"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MenuRepository interface {
	GetAll(ctx context.Context, db *gorm.DB, req *menu.GetRequest) ([]entity.Menu, error)
	GetByRoleID(ctx context.Context, db *gorm.DB, roleID uuid.UUID) ([]entity.Menu, error)
	GetPermission(ctx context.Context, db *gorm.DB, menuID uuid.UUID) ([]entity.MenuPermission, error)
}

type menuRepository struct {
	Log *logrus.Logger
}

func NewMenuRepository(log *logrus.Logger) MenuRepository {
	return &menuRepository{
		Log: log,
	}
}

func (r *menuRepository) GetAll(ctx context.Context, db *gorm.DB, req *menu.GetRequest) ([]entity.Menu, error) {
	query := db.Model(&entity.Menu{}).Where("deleted_at IS NULL")

	var menus []entity.Menu
	if err := query.Find(&menus).Error; err != nil {
		r.Log.Errorf("failed to get menu: %v\n", err)
		return nil, err
	}

	return menus, nil
}

func (r *menuRepository) GetByRoleID(ctx context.Context, db *gorm.DB, roleID uuid.UUID) ([]entity.Menu, error) {
	var menus []entity.Menu
	rows, err := db.
		Raw(`
			SELECT 
				menus.id, 
				menus.name, 
				menus.path, 
				menus.parent_id, 
				menus.icon, 
				menus."order"
			FROM menus
			LEFT JOIN menu_permissions 
				ON menus.id = menu_permissions.menu_id
			LEFT JOIN menu_permission_roles 
				ON menu_permissions.id = menu_permission_roles.menu_permission_id
			WHERE menu_permission_roles.role_id = ?
				AND menus.deleted_at IS NULL
				AND menu_permission_roles.deleted_at IS NULL
				AND menus.is_active = TRUE
				-- OR (menus.parent_id IS NULL AND menus.is_active = TRUE)
			GROUP BY menus.id
			ORDER BY menus."order"
		`, roleID).Rows()
	if err != nil {
		r.Log.Errorf("failed to get menus by role ID: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var menu entity.Menu
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Path, &menu.ParentID, &menu.Icon, &menu.Order); err != nil {
			r.Log.Errorf("failed to scan menu: %v\n", err)
			return nil, err
		}
		menus = append(menus, menu)
	}
	return menus, nil
}

func (r *menuRepository) GetPermission(ctx context.Context, db *gorm.DB, menuID uuid.UUID) ([]entity.MenuPermission, error) {
	var permissions []entity.MenuPermission
	if err := db.Where("menu_id = ? AND deleted_at IS NULL", menuID).Order(`"order"`).Find(&permissions).Error; err != nil {
		r.Log.Errorf("failed to get menu permissions: %v\n", err)
		return nil, err
	}
	return permissions, nil
}
