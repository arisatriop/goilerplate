package dtorequest

type BarCreateRequest struct {
	Code string `json:"code" validate:"required,gte=3" example:"EXP001"`
	Bar  string `json:"bar" validate:"required" example:"First bar"`
}

// BarUpdateRequest replaces a bar's content. Code is optional and, when sent, must equal the
// bar's code: it is the business key and cannot be changed.
type BarUpdateRequest struct {
	Code string `json:"code" validate:"omitempty" example:"EXP001"`
	Bar  string `json:"bar" validate:"required" example:"First bar"`
}

// BarIDParam is the :id path segment of /bars/:id. It is validated as a UUID so a malformed value
// is a 400 here rather than a type error from PostgreSQL.
type BarIDParam struct {
	ID string `params:"id" validate:"required,uuid"`
}

type BarListRequest struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}
