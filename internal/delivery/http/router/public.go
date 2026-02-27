package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

type PublicRouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func (r *PublicRouteRegistry) register(route fiber.Router) {
	auth := route.Group("api/v1/auth")
	auth.Post("/register", r.Wired.Handlers.Auth.Register)
	auth.Post("/login", r.Wired.Handlers.Auth.Login)
	auth.Post("/refresh", r.Wired.Middleware.Auth.AuthenticateRefreshToken(), r.Wired.Handlers.Auth.RefreshToken)
	auth.Post("/logout", r.Wired.Middleware.Auth.Authenticate(), r.Wired.Handlers.Auth.Logout)
	auth.Post("/logout-all", r.Wired.Middleware.Auth.Authenticate(), r.Wired.Handlers.Auth.LogoutAll)

	api := route.Group("api").Use(r.Wired.Middleware.Auth.Authenticate())
	v1 := api.Group("v1")

	r.template(v1)
	r.example(v1)
}

func (r *PublicRouteRegistry) template(v1 fiber.Router) {
	template := v1.Group("templates")
	template.Post("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionTemplateCreate),
		r.Wired.Handlers.Template.Create)

	template.Put("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionTemplateUpdate),
		r.Wired.Handlers.Template.Update)

	template.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionTemplateDelete),
		r.Wired.Handlers.Template.Delete)

	template.Get("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionTemplateList),
		r.Wired.Handlers.Template.List)

	template.Get("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionTemplateDetail),
		r.Wired.Handlers.Template.Get)
}

func (r *PublicRouteRegistry) example(v1 fiber.Router) {
	example := v1.Group("examples")
	example.Post("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionExampleCreate),
		r.Wired.Handlers.Example.Create)

	example.Put("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionExampleUpdate),
		r.Wired.Handlers.Example.Update)

	example.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionExampleDelete),
		r.Wired.Handlers.Example.Delete)

	example.Get("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionExampleList),
		r.Wired.Handlers.Example.List)

	example.Get("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionExampleDetail),
		r.Wired.Handlers.Example.Get)
}
