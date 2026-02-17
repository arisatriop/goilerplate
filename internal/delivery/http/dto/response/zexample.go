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
