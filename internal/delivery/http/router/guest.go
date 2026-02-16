package router

import "github.com/gofiber/fiber/v2"

func (r *RouteRegistry) registerGuestAPI(router fiber.Router) {
	guest := router.Group("/api/v1/guest").Use(r.Wired.Middleware.Store.SetTenantID())

	guest.Get("banners", r.Wired.Handlers.Banner.GetListGuest)
	guest.Get("store", r.Wired.Handlers.Store.GetStoreInfo)
	guest.Get("products", r.Wired.Handlers.Product.GetCategoryListWithProductsByGuest)
	guest.Get("categories", r.Wired.Handlers.Category.GetListByGuest)
}
