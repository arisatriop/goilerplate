package model

import "time"

// UserSession represents the user_sessions table
type UserSession struct {
	ID                 string     `gorm:"type:uuid;primaryKey;column:id"`
	UserID             string     `gorm:"type:uuid;not null;column:user_id"`
	RefreshJTI         string     `gorm:"type:varchar(64);not null;column:refresh_jti"`
	PreviousRefreshJTI *string    `gorm:"type:varchar(64);column:previous_refresh_jti"`
	RotatedAt          *time.Time `gorm:"column:rotated_at"`
	DeviceName         string     `gorm:"type:varchar(255);column:device_name"`
	DeviceType         string     `gorm:"type:varchar(50);column:device_type"`
	DeviceID           string     `gorm:"type:varchar(255);column:device_id"`
	IPAddress          *string    `gorm:"type:inet;column:ip_address"`
	UserAgent          string     `gorm:"type:text;column:user_agent"`
	IsActive           bool       `gorm:"not null;default:true;column:is_active"`
	ExpiresAt          time.Time  `gorm:"not null;column:expires_at"`
	LastUsedAt         time.Time  `gorm:"not null;column:last_used_at"`
	RevokedAt          *time.Time `gorm:"column:revoked_at"`
	RevokedReason      *string    `gorm:"type:varchar(50);column:revoked_reason"`
	CreatedAt          time.Time  `gorm:"not null;column:created_at"`
}

// TableName specifies the table name for UserSession
func (UserSession) TableName() string {
	return "user_sessions"
}
