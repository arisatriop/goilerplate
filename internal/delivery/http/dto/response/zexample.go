package dtoresponse

type ExampleResponse struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Example string `json:"example"`
}

// type ExampleListResponse struct {
// 	Data       []*ExampleResponse  `json:"data"`
// 	Pagination *PaginationResponse `json:"pagination"`
// }

// type PaginationResponse struct {
// 	Page        int   `json:"page"`
// 	Limit       int   `json:"limit"`
// 	Total       int64 `json:"total"`
// 	TotalPages  int   `json:"total_pages"`
// 	HasNext     bool  `json:"has_next"`
// 	HasPrevious bool  `json:"has_previous"`
// }
