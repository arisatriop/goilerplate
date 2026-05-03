package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/bazs"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToBazsFilter(req *dtorequest.BazsListRequest, ctx *fiber.Ctx) *bazs.Filter {
	filter := &bazs.Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}

	return filter
}
