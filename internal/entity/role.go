package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string     `gorm:"column:name;not null;unique;size:50"`
	Order     *int       `gorm:"column:order;default:0"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	CreatedBy string     `gorm:"column:created_by;not null"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	UpdatedBy *string    `gorm:"column:updated_by"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	DeletedBy *string    `gorm:"column:deleted_by"`
}
