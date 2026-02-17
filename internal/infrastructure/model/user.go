package model

import "time"

type User struct {
	ID                  string
	Name                string
	Phone               string
	Email               string
	Avatar              string
	IsActive            bool `gorm:"default:true"`
	PasswordHash        string
	EmailVerified       bool
	EmailVerifiedAt     *time.Time
	PasswordChangedAt   time.Time
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	DeletedAt           *time.Time
	DeletedBy           *string
}

func (User) TableName() string {
	return "users"
}

// UserPermission represents user-specific permission overrides
type UserPermission struct {
	UserID       string `gorm:"column:user_id;primaryKey"`
	PermissionID string `gorm:"column:permission_id;primaryKey"`
	IsGranted    bool   `gorm:"column:is_granted;default:1"`
	CreatedAt    time.Time
	CreatedBy    string
	UpdatedAt    time.Time
	UpdatedBy    string
}

func (UserPermission) TableName() string {
	return "user_permissions"
}
