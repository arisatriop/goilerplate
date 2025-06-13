package role

import (
	"goilerplate/internal/entity"
	"time"

	"github.com/google/uuid"
)

type GetResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Order     *int       `json:"order"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *time.Time `json:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy"`
	DeletedAt *time.Time `json:"deletedAt"`
	DeletedBy *string    `json:"deletedBy"`
}

type GetAllResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Order     *int       `json:"order"`
	CreatedAt time.Time  `json:"-"`
	CreatedBy string     `json:"-"`
	UpdatedAt *time.Time `json:"-"`
	UpdatedBy *string    `json:"-"`
	DeletedAt *time.Time `json:"-"`
	DeletedBy *string    `json:"-"`
}

func ToGet(role *entity.Role) *GetResponse {
	return &GetResponse{
		ID:        role.ID,
		Name:      role.Name,
		Order:     role.Order,
		CreatedAt: role.CreatedAt,
		CreatedBy: role.CreatedBy,
		UpdatedAt: role.UpdatedAt,
		UpdatedBy: role.UpdatedBy,
		DeletedAt: role.DeletedAt,
		DeletedBy: role.DeletedBy,
	}
}

func ToGetAll(role *entity.Role) *GetAllResponse {
	return &GetAllResponse{
		ID:    role.ID,
		Name:  role.Name,
		Order: role.Order,
	}
}
