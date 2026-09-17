package bootstrap

import (
	"testing"

	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// middlewareCount counts the handlers registered with app.Use, as seen by GET requests.
func middlewareCount(t *testing.T, cfg *config.Config) int {
	t.Helper()

	app := NewFiber(cfg)
	stack := app.Stack()
	count := 0
	for i, method := range app.Config().RequestMethods {
		if method != fiber.MethodGet {
			continue
		}
		for _, route := range stack[i] {
			if route.Path == "/" {
				count += len(route.Handlers)
			}
		}
	}
	return count
}

func TestNewFiber_OTelMiddlewareOnlyWhenEnabled(t *testing.T) {
	disabled := middlewareCount(t, &config.Config{})
	enabled := middlewareCount(t, &config.Config{OTel: config.OTel{Enabled: true}})

	assert.Positive(t, disabled, "recover middleware is always registered")
	assert.Equal(t, disabled+1, enabled, "otel.enabled adds exactly the otelfiber middleware")
}

func TestNewGrpcServer_WithAndWithoutOTel(t *testing.T) {
	for _, otelEnabled := range []bool{false, true} {
		server := NewGrpcServer(&config.Config{OTel: config.OTel{Enabled: otelEnabled}})
		assert.NotNil(t, server)
		server.Stop()
	}
}

func TestPartnerRoutesEnabled(t *testing.T) {
	assert.False(t, PartnerRoutesEnabled(&config.Config{}))
	assert.False(t, PartnerRoutesEnabled(&config.Config{Apikeys: map[string]string{}}))
	assert.True(t, PartnerRoutesEnabled(&config.Config{Apikeys: map[string]string{"partner1": "key"}}))
}
