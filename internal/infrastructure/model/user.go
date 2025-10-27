package model

import "time"

type User struct {
	ID                  string
	Name                string
	Email               string
	Avatar              string
	IsActive            bool
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
