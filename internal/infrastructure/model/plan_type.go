package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type PlanType struct {
	ID        uuid.UUID  `gorm:"type:char(36);default:uuid();primaryKey"`
	Code      string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Name      string     `gorm:"type:varchar(100);not null;index"`
	IsActive  bool       `gorm:"type:tinyint(1);not null;default:1;index"`
	CreatedBy string     `gorm:"type:varchar(255);not null;index"`
	UpdatedBy *string    `gorm:"type:varchar(255)"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"index"`
}

func (PlanType) TableName() string {
	return "plan_types"
}

func (pt *PlanType) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	pt.CreatedBy = userID

	return nil
}

func (pt *PlanType) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	if pt.UpdatedBy == nil {
		pt.UpdatedBy = &userID
	} else {
		*pt.UpdatedBy = userID
	}

	return nil
}
