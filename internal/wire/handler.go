package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Example *handler.Example
	Auth    *handler.Auth
	// Future handlers will be added here:
	// UserHandler    *handler.UserHandler
	// OrderHandler   *handler.OrderHandler
	// ProductHandler *handler.ProductHandler
}

// Middleware contains all middleware components
type Middleware struct {
	Auth          *middleware.Auth
	Recover       fiber.Handler
	RequestLogger *middleware.RequestLogger
	// Future middleware will be added here:
	// RateLimit *middleware.RateLimit
	// CORS      *middleware.CORS
	// Logger    *middleware.Logger
}

// WireHandlers creates all HTTP handlers
func WireHandlers(app *bootstrap.App, useCases *UseCases) *Handlers {
	// Create device service
	deviceService := auth.NewDeviceService()

	return &Handlers{
		Auth:    handler.NewAuth(app.Validator, useCases.AuthUC, deviceService),
		Example: handler.NewExample(app.Validator, useCases.ExampleUC),
		// Future handler wiring:
		// UserHandler:    handler.NewUser(app.Validator, useCases.UserUC),
		// OrderHandler:   handler.NewOrder(app.Validator, useCases.OrderUC),
		// ProductHandler: handler.NewProduct(app.Validator, useCases.ProductUC),
	}
}

// WireMiddleware creates all middleware components
func WireMiddleware(app *bootstrap.App, repos *Repositories) *Middleware {
	// Create JWT service
	jwtService := jwt.NewJWTService(
		app.Config.JWT.SecretKey,
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
		app.Config.JWT.Issuer,
		app.Config.JWT.AccessTokenExpiry,
		app.Config.JWT.RefreshTokenExpiry,
	)

	// Create cache service for auth middleware (will be nil if Redis is disabled)
	cacheService := auth.NewCacheService(app.Redis)

	// Create permission service for permission checking (with caching support)
	permissionService := auth.NewPermissionService(repos.AuthRepo, cacheService)

	return &Middleware{
		Auth:          middleware.NewAuth(jwtService, repos.AuthRepo, cacheService, permissionService),
		Recover:       middleware.Recover(),
		RequestLogger: middleware.NewRequestLogger(),
		// Future middleware wiring:
		// RateLimit: middleware.NewRateLimit(),
		// CORS:      middleware.NewCORS(),
		// Logger:    middleware.NewLogger(),
	}
}
