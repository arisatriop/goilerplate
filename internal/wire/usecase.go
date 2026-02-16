package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/banner"
	"goilerplate/internal/domain/category"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/plan"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/domain/productimage"
	"goilerplate/internal/domain/store"
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/domain/zexample2"
	"goilerplate/pkg/jwt"
)

// UseCases contains all use case implementations
type UseCases struct {
	AuthUC         auth.Usecase
	ExampleUC      zexample.Usecase
	ExampleUC2     zexample2.Usecase
	BannerUC       banner.Usecase
	CategoryUC     category.Usecase
	PlanUC         plan.Usecase
	PlantypeUC     plantype.Usecase
	PlantypeRuleUC plantyperule.Usecase
	ProductUC      product.Usecase
	ProductImageUC productimage.Usecase
	StoreUC        store.Usecase
	OrderUC        order.Usecase
	// Future use cases will be added here:
	// UserUC    user.UseCase
	// OrderUC   order.UseCase
	// ProductUC product.UseCase
}

// WireUseCases creates all use case implementations
func WireUseCases(app *bootstrap.App, repos *Repositories, infra *Infrastructure) *UseCases {
	// Create JWT service for auth use case
	jwtService := jwt.NewJWTService(
		app.Config.JWT.SecretKey,
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
		app.Config.JWT.Issuer,
		app.Config.JWT.AccessTokenExpiry,
		app.Config.JWT.RefreshTokenExpiry,
	)

	// Create cache service for auth (will be nil if Redis is disabled)
	cacheService := auth.NewCacheService(app.Redis)

	return &UseCases{
		AuthUC:         auth.NewUseCase(repos.AuthRepo, jwtService, cacheService),
		ExampleUC:      zexample.NewUseCase(repos.ExampleRepo),
		ExampleUC2:     zexample2.NewUseCase(repos.ExampleRepo2),
		PlanUC:         plan.NewUseCase(repos.PlanRepo),
		BannerUC:       banner.NewUseCase(repos.BannerRepo, infra.FilesystemManager),
		CategoryUC:     category.NewUseCase(repos.CategoryRepo),
		PlantypeUC:     plantype.NewUseCase(repos.PlanTypeRepo),
		PlantypeRuleUC: plantyperule.NewUseCase(repos.PlanTypeRuleRepo),
		ProductUC:      product.NewUseCase(repos.ProductRepo),
		ProductImageUC: productimage.NewUseCase(repos.ProductImageRepo, infra.FilesystemManager),
		StoreUC:        store.NewUsecase(repos.StoreRepo),
		OrderUC:        order.NewUseCase(repos.OrderRepo),
		// Future use case wiring:
		// UserUC:    user.NewUseCase(repos.UserRepo),
		// OrderUC:   order.NewUseCase(repos.OrderRepo, repos.ProductRepo),
		// ProductUC: product.NewUseCase(repos.ProductRepo),
	}
}
