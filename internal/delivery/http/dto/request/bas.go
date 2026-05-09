package dtorequest

type BasCreateRequest struct {
	Code string `json:"code" validate:"required"`
	Bas  string `json:"bas" validate:"required"`
}
