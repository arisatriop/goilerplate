package dtorequest

type FooCreateRequest struct {
	Foo string `json:"foo" validate:"required"`
}

type FooUpdateRequest struct {
	Foo string `json:"foo" validate:"required"`
}
