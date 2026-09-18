package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/bar"
	"goilerplate/internal/domain/foo"
)

// UseCases contains all use case implementations
type UseCases struct {
	AuthUC auth.Usecase
	FooUC  foo.Usecase
	BarUC  bar.Usecase
	// Future use cases will be added here:
	// UserUC    user.UseCase
	// OrderUC   order.UseCase
	// ProductUC product.UseCase
}

// WireUseCases creates all use case implementations
func WireUseCases(app *bootstrap.App, repos *Repositories, infra *Infrastructure) *UseCases {
	// The middleware verifies the tokens this use case issues, so both must share one
	// service: a second instance could drift to a different key or issuer.
	sessionService := auth.NewSessionService(repos.AuthRepo, infra.SessionStore)
	permissionService := auth.NewPermissionService(repos.AuthRepo, infra.PermissionCache)

	sessions := auth.SessionExpiry{
		Default:    app.Config.Auth.SessionExpiryOrDefault(),
		RememberMe: app.Config.Auth.RememberMeExpiryOrDefault(),
	}

	return &UseCases{
		AuthUC: auth.NewUseCase(repos.AuthRepo, infra.JWTService, sessionService, permissionService, sessions),
		FooUC:  foo.NewUseCase(repos.FooRepo),
		BarUC:  bar.NewUseCase(repos.BarRepo),
		// Future use cases will be added here:
		// UserUC:    user.NewUseCase(repos.UserRepo),
		// OrderUC:   order.NewUseCase(repos.OrderRepo, repos.ProductRepo),
		// ProductUC: product.NewUseCase(repos.ProductRepo),
	}
}
