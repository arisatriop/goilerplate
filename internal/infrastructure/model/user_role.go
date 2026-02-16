package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	UserID    uuid.UUID `gorm:"type:char(36);primaryKey"`
	RoleID    uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string    `gorm:"type:varchar(255);not null"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
