package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductImage struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;default:uuid()"`
	ProductID   uuid.UUID  `gorm:"column:product_id;not null"`
	FileType    string     `gorm:"column:file_type;not null"`
	FileStorage string     `gorm:"column:file_storage;not null"`
	FileName    string     `gorm:"column:file_name;not null"`
	FilePath    string     `gorm:"column:file_path;not null"`
	FileURL     string     `gorm:"column:file_url;not null"`
	IsPrimary   bool       `gorm:"column:is_primary;not null;default:0"`
	IsActive    bool       `gorm:"column:is_active;not null;default:1"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy   string     `gorm:"column:created_by;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedBy   string     `gorm:"column:updated_by;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
	DeletedBy   *string    `gorm:"column:deleted_by"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

func (p *ProductImage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.CreatedBy = userID
	p.UpdatedBy = userID

	return nil
}

func (p *ProductImage) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	p.UpdatedBy = userID

	return nil
}
