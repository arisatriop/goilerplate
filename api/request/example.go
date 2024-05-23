package request

import (
	"github.com/gofiber/fiber/v2"
)

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

func (p *ExampleImpl) GetReadPayload(ctx *fiber.Ctx) (*ExampleReadPayload, error) {
	return &ExampleReadPayload{
		Search: ctx.FormValue("search"),
		Limit:  ctx.FormValue("limit"),
		Offset: ctx.FormValue("offset"),
	}, nil

}

func (p *ExampleImpl) GetCreatePayload(ctx *fiber.Ctx) (*ExampleCreatePayload, error) {
	return &ExampleCreatePayload{
		Code:    ctx.FormValue("code"),
		Example: ctx.FormValue("example"),
	}, nil
}

func (p *ExampleImpl) GetUpdatePayload(ctx *fiber.Ctx) (*ExampleUpdatePayload, error) {
	return &ExampleUpdatePayload{
		Code:    ctx.FormValue("code"),
		Example: ctx.FormValue("example"),
	}, nil
}
