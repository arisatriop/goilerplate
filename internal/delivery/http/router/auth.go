package router

import (
	"github.com/gofiber/fiber/v2"
)

func (r *RouteRegistry) registerAuthAPI(route fiber.Router) {

	// Create "auth" group
	auth := route.Group("api/v1/auth")

	// Public endpoints (no authentication required)
	auth.Post("/register", r.Wired.Handlers.Auth.Register)
	auth.Post("/login", r.Wired.Handlers.Auth.Login)

	// Refresh token endpoint (requires refresh token, not access token)
	auth.Post("/refresh", r.Wired.Middleware.Auth.AuthenticateRefreshToken(), r.Wired.Handlers.Auth.RefreshToken)

	// Protected endpoints (require access token authentication middleware)
	authProtected := auth.Group("", r.Wired.Middleware.Auth.Authenticate())
	authProtected.Post("/logout", r.Wired.Handlers.Auth.Logout)
	authProtected.Post("/logout-all", r.Wired.Handlers.Auth.LogoutAll)
}
