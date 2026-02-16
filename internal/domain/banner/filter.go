package banner

import (
	"goilerplate/pkg/pagination"

	"github.com/google/uuid"
)

type Filter struct {
	StoreID  *uuid.UUID
	IsActive *bool

	Pagination *pagination.PaginationRequest
}
