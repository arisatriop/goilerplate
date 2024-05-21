package request

import "github.com/gofiber/fiber/v2"

type ExampleImpl struct{}

type ExampleReadPayload struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Search string `json:"search" form:"search"`
}

type ExampleCreatePayload struct {
	Code    string `json:"code" form:"code"`
	Example string `validate:"required" json:"example" form:"example"`
}

type ExampleUpdatePayload struct {
	Code    string `json:"code" form:"code"`
	Example string `validate:"required" json:"example" form:"example"`
}

type IExample interface {
	GetReadPayload(ctx *fiber.Ctx) (*ExampleReadPayload, error)
	GetCreatePayload(ctx *fiber.Ctx) (*ExampleCreatePayload, error)
	GetUpdatePayload(ctx *fiber.Ctx) (*ExampleUpdatePayload, error)
}

func NewExampleRequest() IExample {
	return &ExampleImpl{}
}

func (payload *ExampleImpl) GetReadPayload(ctx *fiber.Ctx) (*ExampleReadPayload, error) {
	panic("Not implement")
}

func (payload *ExampleImpl) GetCreatePayload(ctx *fiber.Ctx) (*ExampleCreatePayload, error) {
	panic("Not implement")
}

func (payload *ExampleImpl) GetUpdatePayload(ctx *fiber.Ctx) (*ExampleUpdatePayload, error) {
	panic("Not implement")
}
