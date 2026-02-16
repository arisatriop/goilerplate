package request

import (
	"goilerplate/internal/domain/order"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
)

func ToOrderFilter(ctx *fiber.Ctx) *order.Filter {
	filter := &order.Filter{}

	// Get StoreID from context
	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	filter.StoreID = storeIDStr

	// Parse optional keyword query parameter
	keyword := ctx.Query("search")
	if keyword != "" {
		filter.Keyword = &keyword
	}

	// Parse pagination parameters
	filter.Pagination = pagination.ParsePagination(ctx)

	return filter
}
