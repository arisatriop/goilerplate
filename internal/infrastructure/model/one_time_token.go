package model

import (
	"time"
)

// OneTimeToken represents the one_time_tokens table
type OneTimeToken struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	UserID    string    `gorm:"type:uuid;not null"`
	TokenType string    `gorm:"type:varchar(50);not null"`
	TokenHash string    `gorm:"type:varchar(255);not null;unique"`
	Attempts  int       `gorm:"not null;default:0"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	IPAddress *string `gorm:"type:inet"`
	UserAgent string  `gorm:"type:text"`
	CreatedAt time.Time
}

// TableName specifies the table name for OneTimeToken
func (OneTimeToken) TableName() string {
	return "one_time_tokens"
}
