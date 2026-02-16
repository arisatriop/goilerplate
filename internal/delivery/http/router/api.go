package router

import (
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) registerPublicAPI(route fiber.Router) {
	auth := route.Group("api/v1/auth")
	auth.Post("/register", r.Wired.Handlers.Auth.Register)
	auth.Post("/login", r.Wired.Handlers.Auth.Login)
	auth.Post("/refresh", r.Wired.Middleware.Auth.AuthenticateRefreshToken(), r.Wired.Handlers.Auth.RefreshToken)
	auth.Post("/logout", r.Wired.Middleware.Auth.Authenticate(), r.Wired.Handlers.Auth.Logout)
	auth.Post("/logout-all", r.Wired.Middleware.Auth.Authenticate(), r.Wired.Handlers.Auth.LogoutAll)

	api := route.Group("api").Use(r.Wired.Middleware.Auth.Authenticate())
	v1 := api.Group("v1")

	r.examplePublic(v1)
}

func (r *RouteRegistry) examplePublic(v1 fiber.Router) {
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
