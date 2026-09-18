package wire

import (
	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/bar"
	"goilerplate/internal/domain/foo"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/password"
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
	strictRevocation := app.Config.Auth.RevocationMode() == config.RevocationStrict
	sessionService := auth.NewSessionService(repos.AuthRepo, infra.SessionStore, strictRevocation)
	permissionService := auth.NewPermissionService(repos.AuthRepo, infra.PermissionCache)

	sessions := auth.SessionExpiry{
		Default:    app.Config.Auth.SessionExpiryOrDefault(),
		RememberMe: app.Config.Auth.RememberMeExpiryOrDefault(),
	}

	return &UseCases{
		AuthUC: auth.NewUseCase(
			repos.AuthRepo, infra.JWTService, sessionService, permissionService,
			sessions, app.Config.Auth.RefreshReuseGraceOrDefault(),
			transaction.NewGormTransaction(app.DB.GDB),
			auth.Lockout{
				MaxAttempts: app.Config.Auth.Lockout.MaxAttemptsOrDefault(),
				Duration:    app.Config.Auth.Lockout.DurationOrDefault(),
			},
			// Same policy as registration; supply a CommonChecker to enable the
			// common-password check (roadmap T3.7).
			password.NewPolicy(nil),
		),
		FooUC: foo.NewUseCase(repos.FooRepo),
		BarUC: bar.NewUseCase(repos.BarRepo),
		// Future use cases will be added here:
		// UserUC:    user.NewUseCase(repos.UserRepo),
		// OrderUC:   order.NewUseCase(repos.OrderRepo, repos.ProductRepo),
		// ProductUC: product.NewUseCase(repos.ProductRepo),
	}
}
