package product

import (
	"goilerplate/pkg/pagination"

	"github.com/google/uuid"
)

type Filter struct {
	ProductID *uuid.UUID
	StoreID   *uuid.UUID
	IsActive  *bool

	Pagination *pagination.PaginationRequest

	CategoryID *uuid.UUID
	Keyword    string
}
