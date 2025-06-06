package zexample

import (
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/helper"
	"time"

	"github.com/google/uuid"
)

type DeleteRequest struct {
	DeletedAt *time.Time `json:"deleted_at" validate:""`
	DeletedBy uuid.UUID  `json:"deleted_by" validate:""`
}

func (r *DeleteRequest) ToDelete(example *entity.Example) {
	now := helper.NowJakarta()
	example.DeletedAt = &now
	example.DeletedBy = &r.DeletedBy
}
