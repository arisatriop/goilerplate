package foo

import (
	"goilerplate/pkg/pagination"
)

type Filter struct {
	Keyword  string
	IsActive *bool

	Pagination *pagination.PaginationRequest
}
