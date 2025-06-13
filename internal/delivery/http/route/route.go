package route

import (
	"goilerplate/internal/delivery/http"
	"goilerplate/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App  *fiber.App
	Auth *middleware.Auth

	ExampleController  http.ExampleController
	AuthController     http.AuthController
	UserController     http.UserController
	MenuController     http.MenuController
	RoleController     http.RoleController
	MenuPermController http.MenuPermissionController
}

func (c *RouteConfig) Setup() {

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

	api := c.App.Group("api").Use(c.Auth.Authenticated())
	v1 := api.Group("v1")
	_ = api.Group("v2")

	examples := v1.Group("examples")
	examples.Post("", c.Auth.Authorized("example:create"), c.ExampleController.Create)
	examples.Get("", c.Auth.Authorized("example:list"), c.ExampleController.GetAll)
	examples.Get(":id", c.Auth.Authorized("example:view"), c.ExampleController.Get)
	examples.Put(":id", c.Auth.Authorized("example:update"), c.ExampleController.Update)
	examples.Delete(":id", c.Auth.Authorized("example:delete"), c.ExampleController.Delete)

	manage := v1.Group("manage")
	manage.Get("/menu-permissions", c.Auth.Authorized("custom.menu_permission:list"), c.MenuPermController.GetAll)

	roles := manage.Group("roles")
	roles.Post("/", c.Auth.Authorized("manage.role:create"), c.RoleController.Create)
	roles.Get("/", c.Auth.Authorized("manage.role:list"), c.RoleController.GetAll)
	roles.Get("/:id", c.Auth.Authorized("manage.role:view"), c.RoleController.Get)
	roles.Put("/:id", c.Auth.Authorized("manage.role:update"), c.RoleController.Update)
	roles.Delete("/:id", c.Auth.Authorized("manage.role:delete"), c.RoleController.Delete)

	users := manage.Group("users")
	users.Post("/", c.Auth.Authorized("manage.user:create"), c.UserController.Create)
	users.Get("/", c.Auth.Authorized("manage.user:list"), c.UserController.GetAll)
	users.Get("/:id", c.Auth.Authorized("manage.user:view"), c.UserController.Get)
	users.Put("/:id", c.Auth.Authorized("manage.user:update"), c.UserController.Update)
	users.Delete("/:id", c.Auth.Authorized("manage.user:delete"), c.UserController.Delete)

	menus := manage.Group("menus")
	menus.Get("", c.Auth.Authorized("manage.menu:list"), c.MenuController.GetAll)
	// menus.Post("", c.Auth.Authorized("manage.menu:create"), c.MenuController.Create)
	// menus.Get(":id", c.Auth.Authorized("manage.menu:view"), c.MenuController.Get)
	// menus.Put(":id", c.Auth.Authorized("manage.menu:update"), c.MenuController.Update)
	// menus.Delete(":id", c.Auth.Authorized("manage.menu:delete"), c.MenuController.Delete)

}

func registerAuthRoutes(c *RouteConfig) {
	auth := c.App.Group("api/v1/auth")
	auth.Post("token", c.AuthController.Token)
	auth.Post("login", c.AuthController.Login)
	auth.Post("logout", c.Auth.Authenticated(), c.AuthController.Logout)
	auth.Get("", c.Auth.Authenticated(), c.AuthController.Me)
}
