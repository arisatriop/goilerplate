package router

import (
	"strings"
	"testing"

	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/wire"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registeredRoutes registers every audience on a fresh app and returns what it exposes. The
// handlers and middleware are zero values: registration only takes method values and builds
// closures, so nothing here needs a database — and nothing is ever called.
func registeredRoutes(t *testing.T) []fiber.Route {
	t.Helper()

	wired := &wire.ApplicationContainer{
		Handlers: &wire.Handlers{
			Auth: &handler.Auth{}, Foo: &handler.Foo{}, Bar: &handler.Bar{}, Upload: &handler.Upload{},
		},
		Middleware: &wire.Middleware{
			Auth:      &middleware.Auth{},
			RateLimit: &middleware.RateLimiter{},
		},
	}

	app := fiber.New()
	(&InternalRouteRegistry{Wired: wired}).register(app)
	(&PartnerRouteRegistry{Wired: wired}).register(app)
	(&PublicRouteRegistry{Wired: wired}).register(app)

	routes := app.GetRoutes(true)
	require.NotEmpty(t, routes)
	return routes
}

// foo is a template whose every method panics. A5 unrouted it from public.go and missed the other
// two audiences, so /internal/foos and /partner/v1/foos kept answering 500. Checking every
// audience at once is what stops the next router file from reintroducing it.
func TestRegister_NoAudienceExposesTheFooTemplate(t *testing.T) {
	for _, route := range registeredRoutes(t) {
		assert.NotContains(t, route.Path, "/foos", "%s %s routes to the foo template", route.Method, route.Path)
	}
}

// The worked example must stay reachable on every audience, or the test above could pass by
// registering nothing.
func TestRegister_EveryAudienceStillExposesBar(t *testing.T) {
	prefixes := map[string]bool{"/internal/bars": false, "/partner/v1/bars": false, "/api/v1/bars": false}

	for _, route := range registeredRoutes(t) {
		for prefix := range prefixes {
			if strings.HasPrefix(route.Path, prefix) {
				prefixes[prefix] = true
			}
		}
	}

	for prefix, found := range prefixes {
		assert.True(t, found, "%s is not registered", prefix)
	}
}
