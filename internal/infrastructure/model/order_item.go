package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type OrderItem struct {
	ID        string          `gorm:"type:char(36);default:uuid();primaryKey"`
	OrderID   string          `gorm:"type:char(36);not null"`
	ProductID string          `gorm:"type:char(36);not null"`
	Name      string          `gorm:"type:varchar(255);not null"`
	Price     decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	Image     *string         `gorm:"type:text"`
	Quantity  int             `gorm:"type:int;not null;default:1"`
	Subtotal  decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	Notes     *string         `gorm:"type:text"`
	CreatedAt time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string          `gorm:"type:char(36);not null;default:system"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	oi.CreatedBy = userID

	return nil
}
