package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/banner"
	"goilerplate/internal/domain/category"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/orderitem"
	"goilerplate/internal/domain/orderstatushistory"
	"goilerplate/internal/domain/plan"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/domain/productcategory"
	"goilerplate/internal/domain/productimage"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/store"
	"goilerplate/internal/domain/subscription"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/domain/zexample2"
	"goilerplate/internal/infrastructure/repository"
)

// Repositories contains all repository implementations
type Repositories struct {
	AuthRepo               auth.Repository
	BannerRepo             banner.Repository
	CategoryRepo           category.Repository
	ExampleRepo            zexample.Repository
	ExampleRepo2           zexample2.Repository
	PlanRepo               plan.Repository
	PlanTypeRepo           plantype.Repository
	PlanTypeRuleRepo       plantyperule.Repository
	ProductCategoryRepo    productcategory.Repository
	ProductImageRepo       productimage.Repository
	ProductRepo            product.Repository
	RoleRepo               role.Repository
	StoreRepo              store.Repository
	SubscriptionRepo       subscription.Repository
	UserRepo               user.Repository
	UserRoleRepo           userrole.Repository
	OrderRepo              order.Repository
	OrderItemRepo          orderitem.Repository
	OrderStatusHistoryRepo orderstatushistory.Repository
}

// WireRepositories creates all repository implementations
func WireRepositories(app *bootstrap.App) *Repositories {
	db := app.DB.GDB
	return &Repositories{
		AuthRepo:               repository.NewAuth(db),
		BannerRepo:             repository.NewBanner(db),
		CategoryRepo:           repository.NewCategory(db),
		ExampleRepo:            repository.NewExample(db),
		ExampleRepo2:           repository.NewExample2(db),
		PlanRepo:               repository.NewPlan(db),
		PlanTypeRepo:           repository.NewPlanType(db),
		PlanTypeRuleRepo:       repository.NewPlanTypeRuleRepository(db),
		ProductCategoryRepo:    repository.NewProductCategory(db),
		ProductImageRepo:       repository.NewProductImage(db),
		ProductRepo:            repository.NewProduct(db),
		RoleRepo:               repository.NewRole(db),
		StoreRepo:              repository.NewStore(db),
		SubscriptionRepo:       repository.NewSubscription(db),
		UserRepo:               repository.NewUser(db),
		UserRoleRepo:           repository.NewUserRole(db),
		OrderRepo:              repository.NewOrder(db),
		OrderItemRepo:          repository.NewOrderItem(db),
		OrderStatusHistoryRepo: repository.NewOrderStatusHistory(db),
	}
}
