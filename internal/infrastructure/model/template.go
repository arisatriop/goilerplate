package model

import (
	"time"
)

type Template struct {
	ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
	Code      string     `gorm:"column:code"`
	Template  string     `gorm:"column:template"`
	IsActive  bool       `gorm:"column:is_active"`
	CreatedBy string     `gorm:"column:created_by"`
	UpdatedBy string     `gorm:"column:updated_by"`
	DeletedBy *string    `gorm:"column:deleted_by"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (Template) TableName() string {
	return "templates"
}
