package user

import (
	"goilerplate/internal/entity"
	"goilerplate/internal/model/role"
	"time"

	"github.com/google/uuid"
)

type GetResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	Email     string                `json:"email"`
	Avatar    string                `json:"avatar"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy string                `json:"createdBy"`
	UpdatedAt *time.Time            `json:"updatedAt"`
	UpdatedBy *string               `json:"updatedBy"`
	DeletedAt *time.Time            `json:"deletedAt"`
	DeletedBy *string               `json:"deletedBy"`
	Role      []role.GetAllResponse `json:"roles"`
}

type GetAllResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	Email     string                `json:"email"`
	Avatar    string                `json:"avatar"`
	UpdatedAt *time.Time            `json:"updatedAt"`
	UpdatedBy *string               `json:"updatedBy"`
	Role      []role.GetAllResponse `json:"roles"`
}

func ToGet(user *entity.User) *GetResponse {
	return &GetResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		CreatedBy: user.CreatedBy,
		UpdatedAt: user.UpdatedAt,
		UpdatedBy: user.UpdatedBy,
		DeletedAt: user.DeletedAt,
		DeletedBy: user.DeletedBy,
	}
}

func ToGetAll(user *entity.User) *GetAllResponse {
	var roles []role.GetAllResponse
	for _, r := range user.Role {
		roles = append(roles, *role.ToGetAll(&r))
	}
	return &GetAllResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
		UpdatedBy: user.UpdatedBy,
		Avatar:    user.Avatar,
		Role:      roles,
	}
}
