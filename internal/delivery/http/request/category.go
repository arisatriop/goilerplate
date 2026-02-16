package request

import (
	"goilerplate/internal/domain/category"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ToCategoryFilterByStore(ctx *fiber.Ctx, paging bool) *category.Filter {
	filter := &category.Filter{}

	if paging {
		filter.Pagination = pagination.ParsePagination(ctx)
	}

	isActive := ctx.Query("isActive")
	if isActive != "" {
		isActive := isActive == "true"
		filter.IsActive = &isActive
	}

	filter.Keyword = ctx.Query("search")

	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parsedUUID, err := uuid.Parse(storeIDStr)
	if err == nil {
		filter.StoreID = &parsedUUID
	}

	return filter
}

func ToCategoryWithProductsFilter(ctx *fiber.Ctx) *category.CategoryWithProductsFilter {
	filter := &category.CategoryWithProductsFilter{}

	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parsedUUID, err := uuid.Parse(storeIDStr)
	if err == nil {
		filter.StoreID = parsedUUID
	}

	parsedCategoryUUID, err := uuid.Parse(ctx.Query("categoryId"))
	if err == nil {
		filter.CategoryID = &parsedCategoryUUID
	}

	filter.Keyword = ctx.Query("search")

	return filter
}
