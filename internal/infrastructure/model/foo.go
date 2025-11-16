package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Foo struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Foo       string
	CreatedAt time.Time  `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string     `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time  `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
	UpdatedBy string     `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time `gorm:"index"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (Foo) TableName() string {
	return "foos"
}

func (f *Foo) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	f.CreatedBy = userID
	f.UpdatedBy = userID

	return nil
}

func (f *Foo) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	f.UpdatedBy = userID

	return nil
}
