package model

import (
	"time"

	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type OrderStatusHistory struct {
	ID        string    `gorm:"type:char(36);default:uuid();primaryKey"`
	OrderID   string    `gorm:"type:char(36);not null"`
	Status    string    `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string    `gorm:"type:char(36);not null;default:system"`
}

func (OrderStatusHistory) TableName() string {
	return "order_status_histories"
}

func (osh *OrderStatusHistory) BeforeCreate(tx *gorm.DB) error {
	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	osh.CreatedBy = userID

	return nil
}
