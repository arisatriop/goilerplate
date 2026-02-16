package router

import (
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) example2(route fiber.Router) {

	example := route.Group("examples")

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
