package order

import "goilerplate/pkg/pagination"

type Filter struct {
	ID         string                        // Optional: Filter by order ID
	StoreID    string                        // Required: Scope to current store
	Keyword    *string                       // Optional: Search by queue number or customer name
	Pagination *pagination.PaginationRequest // Optional: Pagination parameters
}
