package dtorequest

type ZexamplenewCreateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type ZexamplenewUpdateRequest struct {
	Code    string `json:"code" validate:"required,gte=3"`
	Example string `json:"example" validate:"required"`
}

type ZexamplenewListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
