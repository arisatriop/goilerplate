package dtorequest

type ExampleCreateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type ExampleUpdateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type ExampleListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
