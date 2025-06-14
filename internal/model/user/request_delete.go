package user

import (
	"goilerplate/internal/entity"
	"goilerplate/pkg/helper"

	"github.com/google/uuid"
)

type DeleteRequest struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

func (r *DeleteRequest) ToDelete(user *entity.User) {
	now := helper.NowJakarta()
	deletedBy := r.DeletedBy.String()
	user.DeletedAt = &now
	user.DeletedBy = &deletedBy
}
