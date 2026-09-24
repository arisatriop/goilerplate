package router

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/wire"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

type PublicRouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func (r *PublicRouteRegistry) register(route fiber.Router) {
	// Unauthenticated endpoints are limited per IP: they are what an attacker can hammer
	// without credentials.
	auth := route.Group("api/v1/auth")
	auth.Post("/register", r.Wired.Middleware.RateLimit.Auth, r.Wired.Handlers.Auth.Register)
	auth.Post("/login", r.Wired.Middleware.RateLimit.Auth, r.Wired.Handlers.Auth.Login)

	// These already carry a verified token, so they are limited per session. Keyed by IP they
	// would make everyone behind one NAT share a budget, and a single user refreshing in a few
	// tabs could lock the rest out.
	auth.Post("/refresh",
		r.Wired.Middleware.Auth.AuthenticateRefreshToken(),
		r.Wired.Middleware.RateLimit.Session,
		r.Wired.Handlers.Auth.RefreshToken)
	auth.Post("/logout",
		r.Wired.Middleware.Auth.Authenticate(),
		r.Wired.Middleware.RateLimit.Session,
		r.Wired.Handlers.Auth.Logout)
	auth.Post("/logout-all",
		r.Wired.Middleware.Auth.Authenticate(),
		r.Wired.Middleware.RateLimit.Session,
		r.Wired.Handlers.Auth.LogoutAll)

	api := route.Group("api").Use(r.Wired.Middleware.Auth.Authenticate(), r.Wired.Middleware.RateLimit.User)
	v1 := api.Group("v1")

	// Changing a password needs the current one, so it is rate limited per user like any other
	// authenticated route rather than per IP.
	v1.Put("/users/me/password", r.Wired.Handlers.Auth.ChangePassword)

	// A user's own devices. No permission is required: the user ID comes from the token, so
	// these can only ever read or revoke the caller's own sessions.
	v1.Get("/users/me/sessions", r.Wired.Handlers.Auth.ListSessions)
	v1.Delete("/users/me/sessions/:id", r.Wired.Handlers.Auth.RevokeSession)

	// foo is the blank template, not a feature. Every one of its layers is
	// panic("Implement me"), so registering it shipped an endpoint that 500s on a fresh clone —
	// the recover middleware catches the panic, but a template presented as a working route is
	// still a bug report waiting to be filed.
	//
	// Uncomment once the foo layers are implemented, or copy this block for your own domain.
	// The pattern to copy is in .claude/skills/crud-operations/SKILL.md.
	// r.foo(v1)

	r.bar(v1)
}

// foo registers the template domain's routes. Deliberately not called; see register above.
//
//nolint:unused // kept as the shape a new domain copies
func (r *PublicRouteRegistry) foo(v1 fiber.Router) {
	foo := v1.Group("foos")
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
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooGet),
		r.Wired.Handlers.Foo.Get)
}

func (r *PublicRouteRegistry) bar(v1 fiber.Router) {
	bar := v1.Group("bars")
	bar.Post("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBarCreate),
		middleware.RequireIdempotencyKey(), r.Wired.Middleware.Idempotency,
		r.Wired.Handlers.Bar.Create)

	bar.Put("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBarUpdate),
		r.Wired.Handlers.Bar.Update)

	bar.Delete("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBarDelete),
		r.Wired.Handlers.Bar.Delete)

	bar.Get("",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBarList),
		r.Wired.Handlers.Bar.List)

	bar.Get("/:id",
		r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionBarGet),
		r.Wired.Handlers.Bar.Get)
}
