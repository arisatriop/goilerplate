package bootstrap

import (
	"errors"
	"io"
	"net"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// The timeouts exist in every example config; before this they were parsed and then dropped,
// so a server advertising read_timeout: 5s in fact had none.
func TestNewFiber_TimeoutsComeFromConfig(t *testing.T) {
	cfg := baseConfig()
	cfg.Server.ReadTimeout = 3 * time.Second
	cfg.Server.WriteTimeout = 4 * time.Second
	cfg.Server.IdleTimeout = 5 * time.Second

	got := NewFiber(cfg).Config()

	assert.Equal(t, 3*time.Second, got.ReadTimeout)
	assert.Equal(t, 4*time.Second, got.WriteTimeout)
	assert.Equal(t, 5*time.Second, got.IdleTimeout)
}

// Unset must mean "a bounded default", not "no limit": a zero ReadTimeout is what leaves the
// connection open forever, and that is the value a minimal config would otherwise produce.
func TestNewFiber_TimeoutsFallBackToBoundedDefaults(t *testing.T) {
	got := NewFiber(baseConfig()).Config()

	assert.Equal(t, config.DefaultServerReadTimeout, got.ReadTimeout)
	assert.Equal(t, config.DefaultServerWriteTimeout, got.WriteTimeout)
	assert.Equal(t, config.DefaultServerIdleTimeout, got.IdleTimeout)
	assert.NotZero(t, got.ReadTimeout)
}

// Asserting the config field only proves it was passed on. This proves the effect: a client
// that opens a connection and never finishes its request is disconnected by the server.
func TestNewFiber_StalledRequestIsCutOffAtReadTimeout(t *testing.T) {
	cfg := baseConfig()
	cfg.Server.ReadTimeout = 250 * time.Millisecond

	app := NewFiber(cfg)
	app.Get("/probe", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = app.Listener(listener) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// A request that is started and never finished — no blank line, so the server keeps waiting
	// for the rest of the headers.
	_, err = conn.Write([]byte("GET /probe HTTP/1.1\r\nHost: localhost\r\n"))
	require.NoError(t, err)

	// Generous relative to the 250ms timeout so a loaded CI runner does not fail the test; the
	// point is that the wait ends at all. Without a read timeout the server never answers and
	// never hangs up, so this read runs to the deadline and the test fails.
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	answer, err := io.ReadAll(conn)

	require.False(t, errors.Is(err, os.ErrDeadlineExceeded),
		"connection was still open 5s after a 250ms read timeout")
	require.NoError(t, err)
	// fasthttp answers a timed-out request with 408 before hanging up. An empty answer is
	// equally acceptable — what matters is that the connection did not stay open.
	if len(answer) > 0 {
		assert.Contains(t, string(answer), "408")
	}
}
