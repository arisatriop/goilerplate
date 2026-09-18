package bootstrap

import (
	"net/http/httptest"
	"testing"

	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func headersFor(t *testing.T, cfg *config.Config, reqHeaders ...map[string]string) map[string]string {
	t.Helper()

	app := NewFiber(cfg)
	app.Get("/probe", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	req := httptest.NewRequest(fiber.MethodGet, "/probe", nil)
	for _, headers := range reqHeaders {
		for name, value := range headers {
			req.Header.Set(name, value)
		}
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	got := map[string]string{}
	for name := range resp.Header {
		got[name] = resp.Header.Get(name)
	}
	return got
}

func baseConfig() *config.Config {
	return &config.Config{App: config.App{Name: "test"}}
}

func TestNewFiber_SecurityHeadersOnEveryResponse(t *testing.T) {
	headers := headersFor(t, baseConfig())

	assert.Equal(t, "nosniff", headers["X-Content-Type-Options"])
	assert.Equal(t, "DENY", headers["X-Frame-Options"])
	assert.Equal(t, "strict-origin-when-cross-origin", headers["Referrer-Policy"])
}

// HSTS is opt-in, and even then only on HTTPS: telling a browser to refuse plain HTTP from a
// service that is only reachable over plain HTTP would lock everyone out.
func TestNewFiber_HSTSIsOptInAndHTTPSOnly(t *testing.T) {
	// Behind a TLS-terminating proxy the scheme arrives in X-Forwarded-Proto, which is believed
	// only from a trusted peer — so this also exercises the trusted-proxy plumbing.
	const viaProxy = "X-Forwarded-Proto"
	https := map[string]string{viaProxy: "https"}

	cfg := baseConfig()
	cfg.Server.TrustedProxies = []string{"0.0.0.0/32"}

	assert.Empty(t, headersFor(t, cfg, https)["Strict-Transport-Security"],
		"HSTS must be off unless asked for")

	cfg.Server.HSTS = true
	assert.Contains(t, headersFor(t, cfg, https)["Strict-Transport-Security"], "max-age=31536000")

	assert.Empty(t, headersFor(t, cfg)["Strict-Transport-Security"],
		"a plain HTTP request must not be told to stop using HTTP")

	// An untrusted peer claiming https must not be able to switch the header on
	untrusted := baseConfig()
	untrusted.Server.HSTS = true
	assert.Empty(t, headersFor(t, untrusted, https)["Strict-Transport-Security"],
		"a spoofed X-Forwarded-Proto must not be believed")
}
