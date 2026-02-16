package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	auditctx "goilerplate/internal/infrastructure/context"
)

type Store struct {
	ID        uuid.UUID  `gorm:"type:char(36);default:uuid();primaryKey"`
	UserID    uuid.UUID  `gorm:"type:char(36);not null;index"`
	Name      string     `gorm:"type:varchar(255);not null"`
	Desc      *string    `gorm:"type:text"`
	Address   *string    `gorm:"type:text"`
	Phone     *string    `gorm:"type:varchar(50)"`
	Email     *string    `gorm:"type:varchar(100)"`
	WebURL    string     `gorm:"type:text;not null"`
	IsActive  bool       `gorm:"type:tinyint(1);not null;default:1;index"`
	CreatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy string     `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	UpdatedBy string     `gorm:"type:varchar(255);not null"`
	DeletedAt *time.Time `gorm:"index"`
	DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (Store) TableName() string {
	return "stores"
}

func (s *Store) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Set audit fields from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.CreatedBy = userID
	s.UpdatedBy = userID

	return nil
}

func (s *Store) BeforeUpdate(tx *gorm.DB) error {
	// Set updated_by from context
	userID := auditctx.GetUserID(tx.Statement.Context)
	s.UpdatedBy = userID

	return nil
}
