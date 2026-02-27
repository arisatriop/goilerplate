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

	r.template(internal)
	r.example(internal)
}

func (r *InternalRouteRegistry) template(internal fiber.Router) {
	template := internal.Group("templates")
	template.Post("",
		r.Wired.Handlers.Template.Create)

	template.Put("/:id",
		r.Wired.Handlers.Template.Update)

	template.Delete("/:id",
		r.Wired.Handlers.Template.Delete)

	template.Get("",
		r.Wired.Handlers.Template.List)

	template.Get("/:id",
		r.Wired.Handlers.Template.Get)
}

func (r *InternalRouteRegistry) example(internal fiber.Router) {
	example := internal.Group("examples")
	example.Post("",
		r.Wired.Handlers.Example.Create)

	example.Put("/:id",
		r.Wired.Handlers.Example.Update)

	example.Delete("/:id",
		r.Wired.Handlers.Example.Delete)

	example.Get("",
		r.Wired.Handlers.Example.List)

	example.Get("/:id",
		r.Wired.Handlers.Example.Get)
}
