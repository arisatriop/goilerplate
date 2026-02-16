package dtorequest

type Example2CreateRequest struct {
	SomethingField1    string   `json:"somethingField1" validate:"required"`
	SomethingField2    string   `json:"somethingField2" validate:"required"`
	SomethingElseField []string `json:"somethingElse" validate:"required,dive,required"`
}
