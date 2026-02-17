package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type Role struct {
	ID          uuid.UUID  `gorm:"type:char(36);default:UUID();primaryKey"`
	Name        string     `gorm:"type:varchar(100);not null"`
	Slug        string     `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy   *string    `gorm:"type:varchar(255)"`
	UpdatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy   *string    `gorm:"type:varchar(255)"`
	DeletedAt   *time.Time `gorm:"type:timestamp"`
	DeletedBy   *string    `gorm:"type:varchar(255)"`
}

func (Role) TableName() string {
	return "roles"
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	if r.CreatedBy == nil {
		r.CreatedBy = &userID
	}

	return nil
}

func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	if r.UpdatedBy == nil {
		r.UpdatedBy = &userID
	} else {
		*r.UpdatedBy = userID
	}

	return nil
}
