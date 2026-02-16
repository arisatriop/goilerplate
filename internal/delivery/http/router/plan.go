package router

import (
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) plan(route fiber.Router) {

	plan := route.Group("plans")

	plan.Get("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionPlanList),
		r.Wired.Handlers.Plan.List)

}
