package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/example"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToExampleFilter(req *dtorequest.ExampleListRequest, ctx *fiber.Ctx) *example.Filter {
	filter := &example.Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}

	return filter
}
