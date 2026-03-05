package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/foo"

	"github.com/gofiber/fiber/v2"
)

func ToFooFilter(req *dtorequest.FooListRequest, ctx *fiber.Ctx) *foo.Filter {
	panic("Implement me")
}
