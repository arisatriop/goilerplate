package http

import (
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

type Route struct {
	App  *fiber.App
	Auth *middleware.Auth
	Log  *middleware.Log

	ExampleHandler  handler.ExampleHandler
	AuthHandler     handler.AuthHandler
	UserHandler     handler.UserHandler
	MenuHandler     handler.MenuHandler
	RoleHandler     handler.RoleHandler
	MenuPermHandler handler.MenuPermissionHandler
}

func (c *Route) Setup() {

	c.App.Use(middleware.Recover())

	c.App.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to Goilerplate!")
	})
	c.App.Get("/health-check", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Service is healthy"})
	})

	registerAuthRoutes(c)

	privateApi := c.App.Group("private-api")
	_ = privateApi.Group("v1")
	_ = privateApi.Group("v2")

	api := c.App.Group("api").Use(c.Log.IncomingReqestLog()).Use(c.Auth.Authenticated())
	v1 := api.Group("v1")
	_ = api.Group("v2")

	examples := v1.Group("examples")
	examples.Post("", c.Auth.Authorized("example:create"), c.ExampleHandler.Create)
	examples.Get("", c.Auth.Authorized("example:list"), c.ExampleHandler.GetAll)
	examples.Get(":id", c.Auth.Authorized("example:view"), c.ExampleHandler.Get)
	examples.Put(":id", c.Auth.Authorized("example:update"), c.ExampleHandler.Update)
	examples.Delete(":id", c.Auth.Authorized("example:delete"), c.ExampleHandler.Delete)

	manage := v1.Group("manage")
	manage.Get("/menu-permissions", c.Auth.Authorized("custom.menu_permission:list"), c.MenuPermHandler.GetAll)

	roles := manage.Group("roles")
	roles.Post("/", c.Auth.Authorized("manage.role:create"), c.RoleHandler.Create)
	roles.Get("/", c.Auth.Authorized("manage.role:list"), c.RoleHandler.GetAll)
	roles.Get("/:id", c.Auth.Authorized("manage.role:view"), c.RoleHandler.Get)
	roles.Put("/:id", c.Auth.Authorized("manage.role:update"), c.RoleHandler.Update)
	roles.Delete("/:id", c.Auth.Authorized("manage.role:delete"), c.RoleHandler.Delete)

	users := manage.Group("users")
	users.Post("/", c.Auth.Authorized("manage.user:create"), c.UserHandler.Create)
	users.Get("/", c.Auth.Authorized("manage.user:list"), c.UserHandler.GetAll)
	users.Get("/:id", c.Auth.Authorized("manage.user:view"), c.UserHandler.Get)
	users.Put("/:id", c.Auth.Authorized("manage.user:update"), c.UserHandler.Update)
	users.Delete("/:id", c.Auth.Authorized("manage.user:delete"), c.UserHandler.Delete)

	menus := manage.Group("menus")
	menus.Get("", c.Auth.Authorized("manage.menu:list"), c.MenuHandler.GetAll)
	// menus.Post("", c.Auth.Authorized("manage.menu:create"), c.MenuHandler.Create)
	// menus.Get(":id", c.Auth.Authorized("manage.menu:view"), c.MenuHandler.Get)
	// menus.Put(":id", c.Auth.Authorized("manage.menu:update"), c.MenuHandler.Update)
	// menus.Delete(":id", c.Auth.Authorized("manage.menu:delete"), c.MenuHandler.Delete)

}

func registerAuthRoutes(c *Route) {
	auth := c.App.Group("api/v1/auth").Use(c.Log.IncomingReqestLog())
	auth.Post("token", c.AuthHandler.Token)
	auth.Post("login", c.AuthHandler.Login)
	auth.Post("logout", c.Auth.Authenticated(), c.AuthHandler.Logout)
	auth.Get("", c.Auth.Authenticated(), c.AuthHandler.Me)
}
