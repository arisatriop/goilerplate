package role

import (
	"goilerplate/internal/entity"

	"github.com/google/uuid"
)

type UpdateRequest struct {
	ID uuid.UUID `json:"id" validate:""`
	CreateRequest
}

func (r *UpdateRequest) ToUpdate(role *entity.Role) error {

	role.Name = r.Name
	role.Order = r.Order
	role.UpdatedBy = r.UpdatedBy

	return nil
}
