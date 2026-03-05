package dtorequest

type BarCreateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Bar string `json:"bar" validate:"required"`
}

type BarUpdateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Bar string `json:"bar" validate:"required"`
}

type BarListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
