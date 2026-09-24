package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"

	"github.com/gofiber/fiber/v2"
)

type InternalRouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func (r *InternalRouteRegistry) register(route fiber.Router) {
	internal := route.Group("/internal").Use(r.Wired.Middleware.Auth.InternalAuthenticate())

	// foo is the blank template, not a feature: every layer is panic("Implement me"), so routing
	// it here ships an endpoint that answers 500. See the matching note in public.go; copy the
	// pattern from .claude/skills/crud-operations/SKILL.md for your own domain.
	// r.foo(internal)
	r.bar(internal)
}

// foo registers the template domain's routes. Deliberately not called; see register above.
//
//nolint:unused // kept as the shape a new domain copies
func (r *InternalRouteRegistry) foo(internal fiber.Router) {
	foo := internal.Group("foos")
	foo.Post("",
		r.Wired.Handlers.Foo.Create)

	foo.Put("/:id",
		r.Wired.Handlers.Foo.Update)

	foo.Delete("/:id",
		r.Wired.Handlers.Foo.Delete)

	foo.Get("",
		r.Wired.Handlers.Foo.List)

	foo.Get("/:id",
		r.Wired.Handlers.Foo.Get)
}

func (r *InternalRouteRegistry) bar(internal fiber.Router) {
	bar := internal.Group("bars")
	bar.Post("",
		r.Wired.Handlers.Bar.Create)

	bar.Put("/:id",
		r.Wired.Handlers.Bar.Update)

	bar.Delete("/:id",
		r.Wired.Handlers.Bar.Delete)

	bar.Get("",
		r.Wired.Handlers.Bar.List)

	bar.Get("/:id",
		r.Wired.Handlers.Bar.Get)
}
