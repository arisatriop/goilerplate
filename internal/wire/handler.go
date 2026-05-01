package wire

import (
	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/domain/auth"

	"github.com/gofiber/fiber/v2"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth   *handler.Auth
	Foo    *handler.Foo
	Bar    *handler.Bar
	Upload *handler.Upload
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
	RateLimit     *middleware.RateLimiter
	// Future middleware will be added here:
	// CORS   *middleware.CORS
	// Logger *middleware.Logger
}

// WireHandlers creates all HTTP handlers
func WireHandlers(app *bootstrap.App, useCases *UseCases, appServices *ApplicationServices, infrastructure *Infrastructure) *Handlers {
	// Create device service
	deviceService := auth.NewDeviceService()

	return &Handlers{
		Auth:   handler.NewAuth(deviceService, app.Validator, appServices.RegisterSvc, useCases.AuthUC),
		Upload: handler.NewUpload(app.Validator, infrastructure.FilesystemManager, app.Config.FileSystem.MaxFileSize),
		Foo:    handler.NewFoo(app.Validator, useCases.FooUC),
		Bar:    handler.NewBar(app.Validator, useCases.BarUC),
	}
}

// WireMiddleware creates all middleware components
func WireMiddleware(cfg *config.Config, repos *Repositories, infrastructure *Infrastructure) *Middleware {
	// Create permission service for permission checking (with caching support)
	permissionService := auth.NewPermissionService(repos.AuthRepo, infrastructure.AuthCacheService)

	return &Middleware{
		Auth:          middleware.NewAuth(infrastructure.JWTService, repos.AuthRepo, infrastructure.AuthCacheService, permissionService, cfg.Apikeys),
		Recover:       middleware.Recover(),
		RequestLogger: middleware.NewRequestLogger(),
		RateLimit:     middleware.NewRateLimiter(cfg.RateLimit),
		// Future middleware wiring:
		// CORS:   middleware.NewCORS(),
		// Logger: middleware.NewLogger(),
	}
}
