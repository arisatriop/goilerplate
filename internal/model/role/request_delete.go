package role

import (
	"goilerplate/internal/entity"
	"goilerplate/internal/helper"
	"time"

	"github.com/google/uuid"
)

type DeleteRequest struct {
	ID        uuid.UUID  `json:"id" validate:"required,uuid"`
	DeletedAt *time.Time `json:"deletedAt" validate:""`
	DeletedBy uuid.UUID  `json:"deletedBy" validate:""`
}

func (r *DeleteRequest) ToDelete(role *entity.Role) {
	now := helper.NowJakarta()
	uuid := r.DeletedBy.String()
	role.DeletedAt = &now
	role.DeletedBy = &uuid
}
