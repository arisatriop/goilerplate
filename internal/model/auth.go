package model

import (
	"goilerplate/internal/model/menu"

	"github.com/google/uuid"
)

type Auth struct {
	// Login user id
	ID       uuid.UUID
	Name     string
	Username string
	Email    string
	Phone    string
	Avatar   string
}

type MeResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Avatar   string    `json:"avatar"`
	MyRole   []MyRole  `json:"role"`
}

type MyRole struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Menu        []menu.GetAllResponse `json:"menu"`
}
