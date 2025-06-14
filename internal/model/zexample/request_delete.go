package zexample

import (
	"goilerplate/internal/entity"
	"goilerplate/pkg/helper"
	"time"

	"github.com/google/uuid"
)

type DeleteRequest struct {
	DeletedAt *time.Time `json:"deletedAt" validate:""`
	DeletedBy uuid.UUID  `json:"deletedBy" validate:""`
}

func (r *DeleteRequest) ToDelete(example *entity.Example) {
	now := helper.NowJakarta()
	example.DeletedAt = &now
	example.DeletedBy = &r.DeletedBy
}
