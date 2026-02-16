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

	r.example(internal)
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
