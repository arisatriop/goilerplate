package model

import (
	auditctx "goilerplate/internal/infrastructure/context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Banner struct {
	ID          uuid.UUID `gorm:"type:char(36);default:uuid();primaryKey"`
	StoreID     uuid.UUID `gorm:"type:char(36);not null;index"`
	Filetype    string    `gorm:"column:file_type"`
	FileStorage string    `gorm:"column:file_storage"`
	Filename    string    `gorm:"column:file_name"`
	Filepath    string    `gorm:"column:file_path"`
	Fileurl     *string   `gorm:"column:file_url"`
	IsActive    bool      `gorm:"type:tinyint(1);not null;index"`

	CreatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string     `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy string     `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time `gorm:"index"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (Banner) TableName() string {
	return "banners"
}

func (s *Banner) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.CreatedBy = userID
	s.UpdatedBy = userID

	return nil
}

func (s *Banner) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.UpdatedBy = userID

	return nil
}
