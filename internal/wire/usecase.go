package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/example"
	"goilerplate/internal/domain/template"
	"goilerplate/pkg/jwt"
)

// UseCases contains all use case implementations
type UseCases struct {
	AuthUC     auth.Usecase
	TemplateUC template.Usecase
	ExampleUC  example.Usecase
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
		AuthUC:     auth.NewUseCase(repos.AuthRepo, jwtService, cacheService),
		TemplateUC: template.NewUseCase(repos.TemplateRepo),
		ExampleUC:  example.NewUseCase(repos.ExampleRepo),
		// Future use cases will be added here:
		// UserUC:    user.NewUseCase(repos.UserRepo),
		// OrderUC:   order.NewUseCase(repos.OrderRepo, repos.ProductRepo),
		// ProductUC: product.NewUseCase(repos.ProductRepo),
	}
}
