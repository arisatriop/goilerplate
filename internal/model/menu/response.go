package menu

import "github.com/google/uuid"

type GetAllResponse struct {
	ID         uuid.UUID        `json:"id"`
	Name       string           `json:"name"`
	Path       string           `json:"path"`
	Icon       string           `json:"icon"`
	Order      int              `json:"order"`
	IsActive   bool             `json:"isActive"`
	Permission []Permission     `json:"permission"`
	Child      []GetAllResponse `json:"child"`
}

type Permission struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order *int   `json:"order"`
}
