package router

import (
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) stores(route fiber.Router) {
	store := route.Group("stores").Use(r.Wired.Middleware.Store.GetStore())

	store.Get("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionStoreDetail),
		r.Wired.Handlers.Store.GetStoreInfo)

	banners := store.Group("/banners")
	banners.Get("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBannerList),
		r.Wired.Handlers.Banner.GetListByStore)
	banners.Post("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBannerCreate),
		r.Wired.Handlers.Banner.Create)
	banners.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBannerDelete),
		r.Wired.Handlers.Banner.Delete)
	banners.Patch("/:id/toggle-active",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBannerUpdate),
		r.Wired.Handlers.Banner.UpdateToggleActive)
	banners.Post("/upload",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBannerUpload),
		r.Wired.Handlers.Banner.Upload)

	categories := store.Group("/categories")
	categories.Get("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionCategoryList),
		r.Wired.Handlers.Category.GetListByStore)
	categories.Post("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionCategoryCreate),
		r.Wired.Handlers.Category.Create)
	categories.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionCategoryDelete),
		r.Wired.Handlers.Category.SoftDelete)
	categories.Patch("/:id/toggle-active",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionCategoryUpdate),
		r.Wired.Handlers.Category.UpdateToggleActive)

	products := store.Group("/products")
	products.Get("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductList),
		r.Wired.Handlers.Product.GetListWithCategories)
	products.Post("/",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductCreate),
		r.Wired.Handlers.Product.Create)
	products.Get("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductDetail),
		r.Wired.Handlers.Product.GetDetails)
	products.Put("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductUpdate),
		r.Wired.Handlers.Product.Update)
	products.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductDelete),
		r.Wired.Handlers.Product.Delete)
	products.Post("/:id/images",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductImageCreate),
		r.Wired.Handlers.Product.CreateProductImages)
	products.Delete("/:id/images",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductImageDelete),
		r.Wired.Handlers.Product.DeleteProductImagesByProductID)
	products.Post("/:id/categories",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductCategoryCreate),
		r.Wired.Handlers.Product.CreateProductCategory)
	products.Put("/:id/categories",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductCategoryUpdate),
		r.Wired.Handlers.Product.UpdateProductCategory)
	products.Delete("/:id/categories",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductCategoryDelete),
		r.Wired.Handlers.Product.DeleteProductCategoryByProductID)
	products.Delete("/:id/categories/:cid",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductCategoryDelete),
		r.Wired.Handlers.Product.DeleteProductCategoryByID)
	products.Patch("/:id/toggle-active",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductUpdate),
		r.Wired.Handlers.Product.UpdateToggleActive)
	products.Patch("/:id/toggle-available",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductUpdate),
		r.Wired.Handlers.Product.UpdateToggleAvailable)

	productImages := store.Group("/product-images")
	productImages.Post("/uploads",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductImageUpload),
		r.Wired.Handlers.ProdctImage.Upload)
	productImages.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductImageDelete),
		r.Wired.Handlers.Product.DeleteProductImageByID)
	productImages.Patch("/:id/primary",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionProductImageMarkAsPrimary),
		r.Wired.Handlers.Product.MarkImageAsPrimary)

	menus := store.Group("/menus")
	menus.Get("/categories",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionMenuCategoryList),
		r.Wired.Handlers.Category.GetMenuCategoryList)

	menus.Get("/products",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionMenuProductList),
		r.Wired.Handlers.Product.GetMenuProductList)

	orders := store.Group("/orders")
	orders.Get("/",
		// r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionOrderList),
		r.Wired.Handlers.Order.GetList)
	orders.Get("/:id",
		// r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionOrderDetail),
		r.Wired.Handlers.Order.GetDetail)
	orders.Post("/",
		// r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionOrderCreate),
		r.Wired.Handlers.Order.CreateOrder)

}
