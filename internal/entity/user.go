package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email;unique"`
	Password  string     `gorm:"column:password"`
	Token     string     `gorm:"column:token"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	CreatedBy string     `gorm:"column:created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	UpdatedBy *string    `gorm:"column:updated_by"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	DeletedBy *string    `gorm:"column:deleted_by"`
}
