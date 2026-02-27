package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/template"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToTemplateFilter(req *dtorequest.TemplateListRequest, ctx *fiber.Ctx) *template.Filter {
	filter := &template.Filter{
		Keyword:    req.Keyword,
		Pagination: pagination.ParsePagination(ctx),
	}

	return filter
}
