package constants

// Permission constants using resource.action format
// These should match the permission slugs in the database

// Example Resource Permissions
const (
	PermissionExampleList   = "example.list"
	PermissionExampleDetail = "example.detail"
	PermissionExampleCreate = "example.create"
	PermissionExampleUpdate = "example.update"
	PermissionExampleDelete = "example.delete"
)

// Add more resource permissions here as needed
const (
	PermissionBannerList   = "banner.list"
	PermissionBannerDetail = "banner.detail"
	PermissionBannerCreate = "banner.create"
	PermissionBannerUpdate = "banner.update"
	PermissionBannerDelete = "banner.delete"
	PermissionBannerUpload = "banner.upload"
)

const (
	PermissionCategoryList   = "category.list"
	PermissionCategoryDetail = "category.detail"
	PermissionCategoryCreate = "category.create"
	PermissionCategoryUpdate = "category.update"
	PermissionCategoryDelete = "category.delete"
)

const (
	PermissionProductList   = "product.list"
	PermissionProductDetail = "product.detail"
	PermissionProductCreate = "product.create"
	PermissionProductUpdate = "product.update"
	PermissionProductDelete = "product.delete"
)

const (
	PermissionProductImageList          = "productImage.list"
	PermissionProductImageDetail        = "productImage.detail"
	PermissionProductImageCreate        = "productImage.create"
	PermissionProductImageDelete        = "productImage.delete"
	PermissionProductImageMarkAsPrimary = "productImage.update"
	PermissionProductImageUpload        = "productImage.upload"
)

const (
	PermissionProductCategoryList   = "productCategory.list"
	PermissionProductCategoryDetail = "productCategory.detail"
	PermissionProductCategoryCreate = "productCategory.create"
	PermissionProductCategoryUpdate = "productCategory.update"
	PermissionProductCategoryDelete = "productCategory.delete"
)

const (
	PermissionMenuCategoryList = "menu.category.list"
	PermissionMenuProductList  = "menu.product.list"
)

const (
	PermissionOrderList   = "order.list"
	PermissionOrderCreate = "order.create"
)

const (
	PermissionPlanList   = "plan.list"
	PermissionPlanDetail = "plan.detail"
)

const (
	PermissionStoreDetail = "store.detail"
)
