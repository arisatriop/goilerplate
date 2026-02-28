package request

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/template"

	"github.com/gofiber/fiber/v2"
)

func ToTemplateFilter(req *dtorequest.TemplateListRequest, ctx *fiber.Ctx) *template.Filter {
	panic("Implement me")
}
