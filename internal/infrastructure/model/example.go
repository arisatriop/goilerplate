package model

import (
	"time"
)

type Example struct {
	ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
	Code      string     `gorm:"column:code"`
	Example   string     `gorm:"column:example"`
	IsActive  bool       `gorm:"column:is_active"`
	CreatedBy string     `gorm:"column:created_by"`
	UpdatedBy string     `gorm:"column:updated_by"`
	DeletedBy *string    `gorm:"column:deleted_by"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (Example) TableName() string {
	return "examples"
}
