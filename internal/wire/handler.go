package wire

import (
	"strings"
	"time"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/delivery/http/refreshtoken"
	"goilerplate/internal/domain/auth"
	pkgcache "goilerplate/pkg/cache"

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
	Idempotency   fiber.Handler
	// Future middleware will be added here:
	// CORS   *middleware.CORS
	// Logger *middleware.Logger
}

// WireHandlers creates all HTTP handlers
func WireHandlers(app *bootstrap.App, useCases *UseCases, appServices *ApplicationServices, infrastructure *Infrastructure) *Handlers {
	// Create device service
	deviceService := auth.NewDeviceService()

	return &Handlers{
		Auth:   handler.NewAuth(deviceService, app.Validator, appServices.RegisterSvc, useCases.AuthUC, refreshTransport(app.Config)),
		Upload: handler.NewUpload(app.Validator, infrastructure.FilesystemManager, app.Config.FileSystem.MaxFileSize),
		Foo:    handler.NewFoo(app.Validator, useCases.FooUC),
		Bar:    handler.NewBar(app.Validator, useCases.BarUC),
	}
}

// WireMiddleware creates all middleware components
func WireMiddleware(cfg *config.Config, repos *Repositories, infrastructure *Infrastructure) *Middleware {
	strictRevocation := cfg.Auth.RevocationMode() == config.RevocationStrict
	sessionService := auth.NewSessionService(repos.AuthRepo, infrastructure.SessionStore, strictRevocation)
	permissionService := auth.NewPermissionService(repos.AuthRepo, infrastructure.PermissionCache)

	return &Middleware{
		Auth:          middleware.NewAuth(infrastructure.JWTService, repos.AuthRepo, sessionService, permissionService, cfg.Apikeys, cfg.InternalAuth, refreshTransport(cfg)),
		Recover:       middleware.Recover(),
		RequestLogger: middleware.NewRequestLogger(omitBodyPaths(cfg)),
		RateLimit:     middleware.NewRateLimiter(cfg.RateLimit, pkgcache.NewFiberStorage(infrastructure.CacheService.GetClient(), "rl:")),
		Idempotency:   middleware.NewIdempotency(infrastructure.IdempotencyStore, infrastructure.Locker, 24*time.Hour),
		// Future middleware wiring:
		// CORS:   middleware.NewCORS(),
		// Logger: middleware.NewLogger(),
	}
}

func omitBodyPaths(cfg *config.Config) []string {
	if cfg.Log == nil {
		return nil
	}
	return cfg.Log.OmitBodyPaths
}

// refreshTransport translates auth.refresh_transport into the delivery layer's options. The
// cookie is Secure everywhere but a local environment, and the pages allowed to refresh with it
// are the CORS allowlist.
func refreshTransport(cfg *config.Config) *refreshtoken.Transport {
	var origins []string
	if cfg.Server.EnableCORS {
		for _, origin := range strings.Split(cfg.Server.CORS.AllowOrigin, ",") {
			if origin = strings.TrimSpace(origin); origin != "" {
				origins = append(origins, origin)
			}
		}
	}

	return refreshtoken.New(refreshtoken.Options{
		Cookie:         cfg.Auth.UsesRefreshCookie(),
		Secure:         !cfg.IsLocal(),
		SameSite:       cfg.Auth.RefreshCookie.SameSite,
		AllowedOrigins: origins,
	})
}
