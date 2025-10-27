package model

import (
	"time"
)

// UserToken represents the user_tokens table model
type UserToken struct {
	ID        string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string     `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"not null;unique;index"`
	TokenType string     `gorm:"type:varchar(50);not null;index"`
	ExpiresAt time.Time  `gorm:"type:timestamp with time zone;not null;index"`
	UsedAt    *time.Time `gorm:"type:timestamp with time zone;index"`
	IPAddress string     `gorm:"type:inet"`
	UserAgent string     `gorm:"type:text"`
}

// TableName specifies the table name for UserToken
func (UserToken) TableName() string {
	return "user_tokens"
}
