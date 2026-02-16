package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type Plan struct {
	ID             uuid.UUID       `gorm:"type:char(36);default:uuid();primaryKey"`
	PlanTypeID     uuid.UUID       `gorm:"type:char(36);not null;index"`
	DurationInDays int             `gorm:"type:int;not null"`
	Price          decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	IsActive       bool            `gorm:"type:tinyint(1);not null;default:1;index"`
	CreatedBy      string          `gorm:"type:varchar(255);not null;index"`
	UpdatedBy      *string         `gorm:"type:varchar(255)"`
	DeletedBy      *string         `gorm:"type:varchar(255)"`
	CreatedAt      time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time       `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt      *time.Time      `gorm:"index"`
}

func (Plan) TableName() string {
	return "plans"
}

func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.CreatedBy = userID

	return nil
}

func (p *Plan) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	if p.UpdatedBy == nil {
		p.UpdatedBy = &userID
	} else {
		*p.UpdatedBy = userID
	}

	return nil
}
