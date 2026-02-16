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
	Auth        *handler.Auth
	Banner      *handler.Banner
	Category    *handler.Category
	Example     *handler.Example
	Example2    *handler.Example2
	Upload      *handler.Upload
	Product     *handler.Product
	ProdctImage *handler.ProductImage
	Store       *handler.Store
	Order       *handler.Order
	Plan        *handler.Plan
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
	Store         *middleware.Store
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
		Auth:        handler.NewAuth(deviceService, app.Validator, appServices.RegistrationService, useCases.AuthUC),
		Example:     handler.NewExample(app.Validator, useCases.ExampleUC),
		Example2:    handler.NewExample2(app.Validator, appServices.ExpSvc),
		Upload:      handler.NewUpload(app.Validator, infrastructure.FilesystemManager, app.Config.FileSystem.MaxFileSize),
		Banner:      handler.NewBanner(app.Validator, useCases.BannerUC, appServices.BannerService),
		Category:    handler.NewCategory(app.Validator, appServices.CategoryService, useCases.CategoryUC),
		Product:     handler.NewProduct(app.Validator, appServices.ProductService, useCases.ProductUC),
		ProdctImage: handler.NewProductImage(app.Validator, useCases.ProductImageUC),
		Store:       handler.NewStore(useCases.StoreUC),
		Order:       handler.NewOrder(app.Validator, appServices.OrderService, useCases.OrderUC),
		Plan:        handler.NewPlan(useCases.PlantypeUC),
	}
}

// WireMiddleware creates all middleware components
func WireMiddleware(cfg *config.Config, repos *Repositories, infrastructure *Infrastructure) *Middleware {
	// Create permission service for permission checking (with caching support)
	permissionService := auth.NewPermissionService(repos.AuthRepo, infrastructure.AuthCacheService)

	return &Middleware{
		Auth:          middleware.NewAuth(infrastructure.JWTService, repos.AuthRepo, infrastructure.AuthCacheService, permissionService),
		Recover:       middleware.Recover(),
		RequestLogger: middleware.NewRequestLogger(),
		Store:         middleware.NewStore(cfg, repos.StoreRepo),
		// Future middleware wiring:
		// RateLimit: middleware.NewRateLimit(),
		// CORS:      middleware.NewCORS(),
		// Logger:    middleware.NewLogger(),
	}
}
