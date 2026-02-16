package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductCategory struct {
	ProductID  uuid.UUID `gorm:"column:product_id;primaryKey;not null"`
	CategoryID uuid.UUID `gorm:"column:category_id;primaryKey;not null"`
	IsActive   bool      `gorm:"column:is_active;not null;default:1"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy  string    `gorm:"column:created_by;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedBy  string    `gorm:"column:updated_by;not null"`
}

func (ProductCategory) TableName() string {
	return "product_categories"
}

func (p *ProductCategory) BeforeCreate(tx *gorm.DB) error {
	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.CreatedBy = userID
	p.UpdatedBy = userID

	return nil
}

func (p *ProductCategory) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.UpdatedBy = userID

	return nil
}
