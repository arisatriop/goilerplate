package router

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
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
	return registeredRoutesWith(t, &config.Config{})
}

func registeredRoutesWith(t *testing.T, cfg *config.Config) []fiber.Route {
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
	(&PublicRouteRegistry{App: &bootstrap.App{Config: cfg}, Wired: wired}).register(app)

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

// Recovery and verification send email, so their routes must not exist where email is not configured: the
// roadmap's "done when" for F1 is that a deployment without mail exposes nothing half-working.
func TestRegister_RecoveryRoutesFollowAuthEmailEnabled(t *testing.T) {
	recoveryPaths := []string{
		"/api/v1/auth/forgot-password", "/api/v1/auth/reset-password",
		"/api/v1/auth/send-verification-email", "/api/v1/auth/verify-email",
		"/api/v1/users/me/email-change", "/api/v1/users/me/email-change/confirm",
	}

	hasPath := func(routes []fiber.Route, path string) bool {
		for _, route := range routes {
			if route.Method == fiber.MethodPost && route.Path == path {
				return true
			}
		}
		return false
	}

	disabled := registeredRoutesWith(t, &config.Config{})
	enabled := registeredRoutesWith(t, &config.Config{Auth: config.Auth{Email: config.AuthEmail{Enabled: true}}})

	for _, path := range recoveryPaths {
		assert.False(t, hasPath(disabled, path), "%s is registered with auth.email.enabled=false", path)
		assert.True(t, hasPath(enabled, path), "%s is missing with auth.email.enabled=true", path)
	}
}

// loadOpenAPIOperations returns "METHOD /path" of every operation in the committed spec, with the
// names of the security schemes each one accepts.
func loadOpenAPIOperations(t *testing.T) map[string][]string {
	t.Helper()

	raw, err := os.ReadFile("../../../../docs/api/openapi.json")
	require.NoError(t, err)

	var spec struct {
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &spec))

	operations := map[string][]string{}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			schemes := []string{}
			for _, requirement := range operation.Security {
				for scheme := range requirement {
					schemes = append(schemes, scheme)
				}
			}
			operations[strings.ToUpper(method)+" "+path] = schemes
		}
	}
	return operations
}

// A route nobody documented is missing from every tool that imports the spec, and it is easy to add one
// and forget: the spec is generated from annotations, which the compiler does not check.
func TestOpenAPI_MatchesRegisteredRoutes(t *testing.T) {
	cfg := &config.Config{}
	cfg.Auth.Email.Enabled = true
	operations := loadOpenAPIOperations(t)

	// Registered by RouteRegistry.Register, which needs a running application.
	registered := map[string]bool{
		"GET /": true, "GET /livez": true, "GET /health": true,
		"GET /readyz": true, "GET /healthcheck": true, "GET /metrics": true,
	}
	param := regexp.MustCompile(`:([a-zA-Z]+)`)
	for _, route := range registeredRoutesWith(t, cfg) {
		// Fiber registers a HEAD alongside every GET.
		if route.Method == fiber.MethodHead {
			continue
		}
		registered[route.Method+" "+param.ReplaceAllString(route.Path, "{$1}")] = true
	}

	for route := range registered {
		assert.Contains(t, operations, route, "route is not in docs/api/openapi.json; annotate it and run make openapi")
	}
	for operation := range operations {
		assert.True(t, registered[operation], "%s is in docs/api/openapi.json but no such route exists", operation)
	}
}

func TestOpenAPI_EveryAudienceDeclaresItsCredential(t *testing.T) {
	public := map[string]bool{
		"POST /api/v1/auth/register": true, "POST /api/v1/auth/login": true,
		"POST /api/v1/auth/forgot-password": true, "POST /api/v1/auth/reset-password": true,
		"POST /api/v1/auth/send-verification-email": true, "POST /api/v1/auth/verify-email": true,
	}

	for operation, schemes := range loadOpenAPIOperations(t) {
		path := strings.SplitN(operation, " ", 2)[1]
		switch {
		case public[operation]:
			assert.Empty(t, schemes, "%s needs no credential", operation)
		case strings.HasPrefix(path, "/api/"):
			assert.Contains(t, schemes, "BearerAuth", operation)
		case strings.HasPrefix(path, "/internal/"):
			assert.Equal(t, []string{"InternalSecret"}, schemes, operation)
		case strings.HasPrefix(path, "/partner/"):
			assert.Equal(t, []string{"ApiKeyAuth"}, schemes, operation)
		default:
			assert.Empty(t, schemes, "%s is a public probe", operation)
		}
	}
}
