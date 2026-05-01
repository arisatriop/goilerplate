package dtorequest

type FooCreateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Bar  string `json:"bar" validate:"required"`
}

type FooUpdateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Bar  string `json:"bar" validate:"required"`
}

type FooListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
