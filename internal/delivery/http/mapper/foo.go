package mapper

import (
	"goilerplate/internal/domain/foo"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"

	dtoresponse "goilerplate/internal/delivery/http/dto/response"
)

func ToFooFilter(ctx *fiber.Ctx, paginate bool) *foo.Filter {
	filter := &foo.Filter{}

	if paginate {
		filter.Pagination = pagination.ParsePagination(ctx)
	}

	filter.Keyword = ctx.Query("keyword")

	return filter
}

func ToFooResponse(foo *foo.Foo) dtoresponse.FooResponse {
	return dtoresponse.FooResponse{
		ID:  foo.ID,
		Foo: foo.Foo,
	}
}
