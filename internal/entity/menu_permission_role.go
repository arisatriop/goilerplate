package entity

import (
	"time"

	"github.com/google/uuid"
)

type MenuPermissionRole struct {
	ID               uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	MenuPermissionID uuid.UUID  `gorm:"column:menu_permission_id"`
	RoleID           uuid.UUID  `gorm:"column:role_id"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	CreatedBy        string     `gorm:"column:created_by;not null"`
	UpdatedAt        *time.Time `gorm:"column:updated_at"`
	UpdatedBy        *string    `gorm:"column:updated_by"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	DeletedBy        *string    `gorm:"column:deleted_by"`
}

func (e *MenuPermissionRole) TableName() string {
	return "menu_permission_roles"
}
