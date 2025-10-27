package zexample

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

type Filter struct {
	Keyword string
	Code    string

	Pagination *pagination.PaginationRequest
}

func ParseFromRequest(req *dtorequest.ExampleListRequest, ctx *fiber.Ctx) *Filter {
	return &Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}
}
