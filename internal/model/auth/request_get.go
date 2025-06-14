package auth

import (
	"github.com/google/uuid"
)

type User struct {
	// Login user id
	ID       uuid.UUID
	Name     string
	Username string
	Email    string
	Phone    string
	Avatar   string
}

type GetResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Avatar   string    `json:"avatar"`
	Role     []Role    `json:"role"`
}
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Menu        []Menu    `json:"menu"`
}

type Menu struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Icon     string    `json:"icon"`
	Order    int       `json:"order"`
	IsActive bool      `json:"isActive"`
	Child    []Menu    `json:"child"`
}
