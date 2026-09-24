package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"

	"github.com/gofiber/fiber/v2"
)

type PartnerRouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func (r *PartnerRouteRegistry) register(route fiber.Router) {
	partner := route.Group("partner").Use(r.Wired.Middleware.Auth.PartnerAuthenticate(), r.Wired.Middleware.RateLimit.Partner)
	v1 := partner.Group("v1")

	// foo is the blank template, not a feature: every layer is panic("Implement me"), so routing
	// it here ships an endpoint that answers 500. See the matching note in public.go; copy the
	// pattern from .claude/skills/crud-operations/SKILL.md for your own domain.
	// r.foo(v1)
	r.bar(v1)
}

// foo registers the template domain's routes. Deliberately not called; see register above.
//
//nolint:unused // kept as the shape a new domain copies
func (r *PartnerRouteRegistry) foo(v1 fiber.Router) {
	foo := v1.Group("foos")
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

func (r *PartnerRouteRegistry) bar(v1 fiber.Router) {
	bar := v1.Group("bars")
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
