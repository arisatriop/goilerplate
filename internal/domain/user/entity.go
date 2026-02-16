package user

import (
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Phone        string
	Email        string
	Avatar       string
	IsActive     bool
	PasswordHash string
}

func (u *User) HashPassword() {
	hasPassword, _ := utils.HashPassword(u.PasswordHash)
	u.PasswordHash = hasPassword
}
