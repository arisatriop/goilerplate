package entity

import (
	"time"

	"github.com/google/uuid"
)

type Menu struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name       string     `gorm:"column:name"`
	Path       string     `gorm:"column:path"`
	Permission string     `gorm:"column:permission"`
	ParentID   *uuid.UUID `gorm:"type:uuid;default:null;column:parent_id"`
	Icon       *string    `gorm:"column:icon"`
	Order      *int       `gorm:"column:order"`
	IsActive   bool       `gorm:"column:is_active"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	CreatedBy  string     `gorm:"type:uuid"`
	UpdatedAt  *time.Time `gorm:"column:updated_at"`
	UpdatedBy  *string    `gorm:"type:uuid"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
	DeletedBy  *string    `gorm:"type:uuid"`
}

func (e *Menu) TableName() string {
	return "menus"
}
