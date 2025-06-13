package user

import (
	"goilerplate/internal/entity"

	"github.com/google/uuid"
)

type UpdateRequest struct {
	ID       uuid.UUID
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	UpdateBy string `json:"updateBy"`
}

func (r *UpdateRequest) ToUpdate(user *entity.User) {
	user.Name = r.Name
	user.Email = r.Email
	user.UpdatedBy = &r.UpdateBy
	if r.Password != "" {
		user.Password = r.Password
	}
	if r.UpdateBy != "" {
		user.CreatedBy = r.UpdateBy
	}
}
