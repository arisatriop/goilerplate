package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type Order struct {
	ID            string          `gorm:"type:char(36);default:uuid();primaryKey"`
	OrderNumber   string          `gorm:"type:varchar(50);not null;uniqueIndex"`
	QueueNumber   *string         `gorm:"type:varchar(20);index:idx_orders_queue_number"`
	TableNumber   *string         `gorm:"type:varchar(20)"`
	OrderType     string          `gorm:"type:varchar(20);not null;comment:e.g., dine-in, takeout"`
	OrderStatus   string          `gorm:"type:varchar(20);not null;default:processing;comment:e.g., processing, completed, cancelled"`
	Notes         *string         `gorm:"type:text"`
	Amount        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00"`
	StoreID       string          `gorm:"type:char(36);not null"`
	CustomerName  *string         `gorm:"type:varchar(255)"`
	CustomerPhone *string         `gorm:"type:varchar(20)"`
	CustomerEmail *string         `gorm:"type:varchar(255)"`
	CreatedAt     time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;index:idx_orders_created_at"`
	CreatedBy     string          `gorm:"type:char(36);not null;default:system"`
	UpdatedAt     time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy     *string         `gorm:"type:char(36)"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	o.CreatedBy = userID
	o.UpdatedBy = &userID

	return nil
}

func (o *Order) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	o.UpdatedBy = &userID

	return nil
}
