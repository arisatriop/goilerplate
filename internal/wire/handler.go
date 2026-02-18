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
	Auth    *handler.Auth
	Example *handler.Example
	Upload  *handler.Upload
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
func WireHandlers(app *bootstrap.App, useCases *UseCases, appServices *ApplicationServices, infrastructure *Infrastructure) *Handlers {
	// Create device service
	deviceService := auth.NewDeviceService()

	return &Handlers{
		Auth:    handler.NewAuth(deviceService, app.Validator, appServices.RegisterSvc, useCases.AuthUC),
		Example: handler.NewExample(app.Validator, useCases.ExampleUC),
		Upload:  handler.NewUpload(app.Validator, infrastructure.FilesystemManager, app.Config.FileSystem.MaxFileSize),
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
		// Future middleware wiring:
		// RateLimit: middleware.NewRateLimit(),
		// CORS:      middleware.NewCORS(),
		// Logger:    middleware.NewLogger(),
	}
}
