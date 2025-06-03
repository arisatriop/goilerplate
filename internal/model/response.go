package model

type Response struct {
	Message    string                `json:"message"`
	Data       any                   `json:"data,omitempty"`
	Pagination *ResponsePageMetadata `json:"pagination,omitempty"`
}

type ResponsePageMetadata struct {
	Page      int    `json:"page"`
	Size      int    `json:"size"`
	TotalItem int64  `json:"total_item"`
	TotalPage int64  `json:"total_page"`
	Next      string `json:"next"`
	Previous  string `json:"previous"`
}
