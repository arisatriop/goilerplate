package model

type RequestPagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
