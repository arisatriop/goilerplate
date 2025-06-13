package role

import (
	"goilerplate/internal/entity"
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Name       string     `json:"name" validate:"required,max=100"`
	Order      *int       `json:"order" validate:"omitempty,min=0"`
	CreatedAt  time.Time  `json:"CreatedAt"`
	CreatedBy  string     `json:"CreatedBy"`
	UpdatedAt  *time.Time `json:"UpdatedAt"`
	UpdatedBy  *string    `json:"UpdatedBy"`
	DeletedAt  *time.Time `json:"DeletedAt"`
	DeletedBy  *string    `json:"DeletedBy"`
	Permission []string   `json:"permission" validate:"dive,required,uuid"`
}

func (r *CreateRequest) ToCreate() *entity.Role {
	return &entity.Role{
		Name:      r.Name,
		Order:     r.Order,
		CreatedBy: r.CreatedBy,
		UpdatedBy: &r.CreatedBy,
	}
}

func (r *CreateRequest) ToCreatePermission(id uuid.UUID) []entity.MenuPermissionRole {
	var permissions []entity.MenuPermissionRole
	for _, perm := range r.Permission {
		permissions = append(permissions, entity.MenuPermissionRole{
			RoleID:           id,
			MenuPermissionID: uuid.MustParse(perm),
			CreatedBy:        *r.UpdatedBy,
			UpdatedBy:        r.UpdatedBy,
		})
	}
	return permissions
}
