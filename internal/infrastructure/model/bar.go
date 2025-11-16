package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bar struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Bar       string
	CreatedAt time.Time  `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string     `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time  `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
	UpdatedBy string     `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time `gorm:"index"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (Bar) TableName() string {
	return "bars"
}

func (b *Bar) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	b.CreatedBy = userID
	b.UpdatedBy = userID

	return nil
}

func (b *Bar) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	b.UpdatedBy = userID

	return nil
}
