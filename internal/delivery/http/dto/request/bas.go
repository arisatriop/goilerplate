package dtorequest

type BasCreateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Name string `json:"name" validate:"required"`
}

type BasUpdateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Name string `json:"name" validate:"required"`
}

type BasListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
