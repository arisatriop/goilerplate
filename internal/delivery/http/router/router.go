package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"

	"github.com/gofiber/fiber/v2"
)

type RouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func NewRouteRegistry(app *bootstrap.App, wired *wire.ApplicationContainer) *RouteRegistry {
	return &RouteRegistry{
		App:   app,
		Wired: wired,
	}
}

// Register sets up all the routes and middleware for the application.
func (r *RouteRegistry) Register() {
	http := r.App.WebServer.Use(r.Wired.Middleware.RequestLogger.LogRequest())
	http.Use(r.Wired.Middleware.Recover)

	http.Get("/", r.index)
	http.Get("/health", r.healthCheck)

	r.registerAuthAPI(http)
	r.registerPublicAPI(http)
	r.registerInternalAPI(http)
}

// registerPublicAPI sets up the public API routes.
func (r *RouteRegistry) registerPublicAPI(router fiber.Router) {
	v1 := router.Group("api/v1").Use(r.Wired.Middleware.Auth.Authenticate())

	r.example(v1)
}

// registerInternalAPI sets up the internal API routes.
func (r *RouteRegistry) registerInternalAPI(router fiber.Router) {
	v1 := router.Group("internal/v1").Use(r.Wired.Middleware.Auth.InternalAuthenticate())

	_ = v1
}
