package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"
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

	r.registerInternalAPI(http)
	r.registerPartnerAPI(http)
	r.registerPublicAPI(http)
}
