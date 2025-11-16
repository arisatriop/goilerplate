package router

import (
	"goilerplate/internal/constants"

	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) foo(route fiber.Router) {

	foo := route.Group("foo")

	foo.Post("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooCreate),
		r.Wired.Handlers.Foo.Create)

	foo.Put("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooUpdate),
		r.Wired.Handlers.Foo.Update)

	foo.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooDelete),
		r.Wired.Handlers.Foo.Delete)

	foo.Get("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooList),
		r.Wired.Handlers.Foo.List)

	foo.Get("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooDetail),
		r.Wired.Handlers.Foo.GetByID)

}
