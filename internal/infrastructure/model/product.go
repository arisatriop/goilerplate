package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID       `gorm:"type:char(36);default:uuid();primaryKey"`
	Name        string          `gorm:"type:text;not null"`
	Description *string         `gorm:"type:text"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null;index"`
	Images      *string         `gorm:"type:text"`
	StoreID     uuid.UUID       `gorm:"type:char(36);not null;index"`
	IsAvailable bool            `gorm:"type:tinyint(1);not null;index"`
	IsActive    bool            `gorm:"type:tinyint(1);not null;index"`

	CreatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string     `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy string     `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time `gorm:"index"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.CreatedBy = userID
	p.UpdatedBy = userID

	return nil
}

func (p *Product) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.UpdatedBy = userID

	return nil
}
