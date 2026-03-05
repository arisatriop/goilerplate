package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/bar"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToBarFilter(req *dtorequest.BarListRequest, ctx *fiber.Ctx) *bar.Filter {
	filter := &bar.Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}

	return filter
}
