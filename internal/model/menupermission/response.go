package menupermission

import (
	"goilerplate/internal/entity"
	"time"

	"github.com/google/uuid"
)

type GetAllResponse struct {
	ID         uuid.UUID  `json:"id"`
	Menu       string     `json:"menu"`
	Permission string     `json:"permission"`
	UpdatedAt  *time.Time `json:"updatedAt"`
	UpdatedBy  *string    `json:"updatedBy"`
}

func ToGetAll(menuperm *entity.MenuPermission) *GetAllResponse {
	return &GetAllResponse{
		ID:         menuperm.ID,
		Menu:       menuperm.Menu.Name,
		Permission: menuperm.Permission,
		UpdatedAt:  menuperm.UpdatedAt,
		UpdatedBy:  menuperm.UpdatedBy,
	}
}
