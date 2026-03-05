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

	r.foo(internal)
	r.bar(internal)
}

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
