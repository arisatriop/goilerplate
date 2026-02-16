package wire

import (
	"goilerplate/internal/application/banner"
	"goilerplate/internal/application/category"
	"goilerplate/internal/application/expapp"
	"goilerplate/internal/application/orderapp"
	"goilerplate/internal/application/productapp"
	"goilerplate/internal/application/registration"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/infrastructure/transaction"
)

// ApplicationServices contains all application services for multi-domain orchestration
type ApplicationServices struct {
	ExpSvc              expapp.ApplicationService
	RegistrationService registration.ApplicationService
	BannerService       banner.ApplicationService
	CategoryService     category.ApplicationService
	ProductService      productapp.ApplicationService
	OrderService        orderapp.ApplicationService
}

func WireApplicationServices(app *bootstrap.App, repos *Repositories, usecases *UseCases, infrastructure *Infrastructure) *ApplicationServices {
	txManager := transaction.NewGormTransaction(app.DB.GDB)

	return &ApplicationServices{
		ExpSvc: expapp.NewApplicationService(
			txManager,
			usecases.ExampleUC,
			repos.ExampleRepo,
			repos.ExampleRepo2,
		),
		RegistrationService: registration.NewApplicationService(
			app.Config,
			txManager,
			usecases.PlanUC,
			repos.UserRepo,
			repos.RoleRepo,
			repos.UserRoleRepo,
			repos.StoreRepo,
			repos.PlanTypeRepo,
			repos.PlanRepo,
			repos.SubscriptionRepo,
		),
		BannerService: banner.NewApplicationService(
			txManager,
			infrastructure.FilesystemManager,
			usecases.BannerUC,
			repos.BannerRepo,
			usecases.PlantypeUC,
			repos.PlanTypeRepo,
			usecases.PlantypeRuleUC,
			repos.PlanTypeRuleRepo,
		),
		CategoryService: category.NewApplicationService(
			txManager,
			infrastructure.CacheService,
			repos.CategoryRepo,
			usecases.PlantypeUC,
			repos.PlanTypeRepo,
			usecases.PlantypeRuleUC,
			repos.PlanTypeRuleRepo,
		),
		ProductService: productapp.NewApplicationService(
			txManager,
			infrastructure.CacheService,
			repos.CategoryRepo,
			usecases.PlantypeUC,
			repos.PlanTypeRepo,
			usecases.PlantypeRuleUC,
			repos.PlanTypeRuleRepo,
			repos.ProductCategoryRepo,
			repos.ProductImageRepo,
			repos.ProductRepo,
		),
		OrderService: orderapp.NewApplicationService(
			txManager,
			usecases.OrderUC,
			repos.OrderRepo,
			repos.OrderItemRepo,
			repos.OrderStatusHistoryRepo,
		),
	}
}
