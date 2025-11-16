package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserToken represents the user_tokens table model
type UserToken struct {
	ID        string     `gorm:"primaryKey;type:char(36);default:uuid()"`
	UserID    string     `gorm:"type:char(36);not null;index"`
	TokenHash string     `gorm:"type:varchar(255);not null;unique;index"`
	TokenType string     `gorm:"type:varchar(50);not null;index"`
	ExpiresAt time.Time  `gorm:"type:datetime(3);not null;index"`
	UsedAt    *time.Time `gorm:"type:datetime(3);index"`
	IPAddress string     `gorm:"type:varchar(45)"`
	UserAgent string     `gorm:"type:text"`
}

// TableName specifies the table name for UserToken
func (UserToken) TableName() string {
	return "user_tokens"
}

func (ut *UserToken) BeforeCreate(tx *gorm.DB) error {
	if ut.ID == "" {
		ut.ID = uuid.NewString()
	}
	return nil
}
