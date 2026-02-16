package zexample2

import (
	"goilerplate/pkg/pagination"
)

type Filter struct {
	Keyword string
	Code    string

	Pagination *pagination.PaginationRequest
}
