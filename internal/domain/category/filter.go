package category

import (
	"goilerplate/pkg/pagination"

	"github.com/google/uuid"
)

type Filter struct {
	StoreID  *uuid.UUID
	IsActive *bool
	Keyword  string

	Pagination *pagination.PaginationRequest
}

type CategoryWithProductsFilter struct {
	StoreID    uuid.UUID
	CategoryID *uuid.UUID

	Keyword string
}
