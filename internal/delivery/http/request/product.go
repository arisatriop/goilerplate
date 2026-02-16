package request

import (
	"goilerplate/internal/domain/product"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ToProductFilter(ctx *fiber.Ctx, paging bool) *product.Filter {

	filter := &product.Filter{
		Keyword: ctx.Query("search"),
	}

	if paging {
		filter.Pagination = pagination.ParsePagination(ctx)
	}

	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parsedUUID, err := uuid.Parse(storeIDStr)
	if err == nil {
		filter.StoreID = &parsedUUID
	}

	categoryIDstr := ctx.Query("categoryId")
	if categoryIDstr != "" {
		parseCategoryID, err := uuid.Parse(categoryIDstr)
		if err == nil {
			filter.CategoryID = &parseCategoryID
		}
	}

	return filter
}
