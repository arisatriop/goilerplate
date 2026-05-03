package dtorequest

type BazsCreateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Name string `json:"name" validate:"required"`
}

type BazsUpdateRequest struct {
	Code string `json:"code" validate:"required,gte=3"`
	Name string `json:"name" validate:"required"`
}

type BazsListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
