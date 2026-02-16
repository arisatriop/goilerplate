package request

import (
	"goilerplate/internal/domain/banner"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ToBannerFilterByStore(ctx *fiber.Ctx, paging bool) *banner.Filter {

	filter := &banner.Filter{}

	if paging {
		filter.Pagination = pagination.ParsePagination(ctx)
	}

	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parsedUUID, err := uuid.Parse(storeIDStr)
	if err == nil {
		filter.StoreID = &parsedUUID
	}

	return filter
}
