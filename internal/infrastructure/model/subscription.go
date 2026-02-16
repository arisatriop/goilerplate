package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type Subscription struct {
	ID        uuid.UUID       `gorm:"type:char(36);default:uuid();primaryKey"`
	StoreID   uuid.UUID       `gorm:"type:char(36);not null;index"`
	PlanID    uuid.UUID       `gorm:"type:char(36);not null;index"`
	StartDate *time.Time      `gorm:"type:timestamp;index"`
	EndDate   *time.Time      `gorm:"type:timestamp"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	Status    string          `gorm:"type:varchar(50);not null;index"`
	IsActive  bool            `gorm:"type:tinyint(1);not null;default:1;index"`
	CreatedAt time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string          `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy string          `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time      `gorm:"index"`
	DeletedBy *string         `gorm:"type:varchar(255)"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.CreatedBy = userID
	s.UpdatedBy = userID

	return nil
}

func (s *Subscription) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.UpdatedBy = userID

	return nil
}
