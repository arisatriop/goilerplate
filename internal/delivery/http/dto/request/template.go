package dtorequest

type TemplateCreateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type TemplateUpdateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type TemplateListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
