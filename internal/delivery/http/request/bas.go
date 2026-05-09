package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/bas"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToBasFilter(req *dtorequest.BasListRequest, ctx *fiber.Ctx) *bas.Filter {
	return &bas.Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}
}
