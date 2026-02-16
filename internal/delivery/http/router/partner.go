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
	partner := route.Group("partner").Use(r.Wired.Middleware.Auth.PartnerAuthenticate())
	v1 := partner.Group("v1")

	r.example(v1)
}

func (r *PartnerRouteRegistry) example(v1 fiber.Router) {
	example := v1.Group("examples")
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
