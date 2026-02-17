package model

import "time"

type UserSession struct {
	ID               string    `gorm:"primaryKey;column:id"`
	UserID           string    `gorm:"column:user_id;not null"`
	RefreshTokenHash string    `gorm:"column:refresh_token_hash;not null;unique"`
	DeviceName       string    `gorm:"column:device_name"`
	DeviceType       string    `gorm:"column:device_type"`
	DeviceID         string    `gorm:"column:device_id"`
	IPAddress        string    `gorm:"column:ip_address"`
	UserAgent        string    `gorm:"column:user_agent"`
	Location         string    `gorm:"column:location"`
	IsActive         bool      `gorm:"column:is_active;not null;default:true"`
	ExpiresAt        time.Time `gorm:"column:expires_at;not null"`
	LastUsedAt       time.Time `gorm:"column:last_used_at;not null"`
}

// TableName specifies the table name for UserSession
func (UserSession) TableName() string {
	return "user_sessions"
}
