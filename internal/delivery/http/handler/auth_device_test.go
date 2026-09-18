package handler

import (
	"net/http/httptest"
	"testing"

	"goilerplate/internal/domain/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureDeviceRequest runs one request through an app configured like the real one and returns
// what the handler would hand to the auth domain.
func captureDeviceRequest(t *testing.T, trustedProxies []string, headers map[string]string) auth.DeviceRequest {
	t.Helper()

	var got auth.DeviceRequest
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})
	app.Get("/", func(ctx *fiber.Ctx) error {
		got = newDeviceRequest(ctx)
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	_, err := app.Test(req)
	require.NoError(t, err)

	return got
}

func TestNewDeviceRequest_CopiesRequestAttributes(t *testing.T) {
	got := captureDeviceRequest(t, nil, map[string]string{
		"User-Agent":      "test-agent",
		"Accept-Language": "id-ID",
		"Accept-Encoding": "gzip",
	})

	assert.Equal(t, "test-agent", got.UserAgent)
	assert.Equal(t, "id-ID", got.AcceptLanguage)
	assert.Equal(t, "gzip", got.AcceptEncoding)
	assert.NotEmpty(t, got.ClientIP)
}

// The point of the trusted-proxy list: a client that sends its own X-Forwarded-For must not be
// able to choose the IP it is recorded and rate limited under. Without this it could rotate the
// header to walk around every per-IP limit and leave a false trail in user_sessions.
func TestNewDeviceRequest_IgnoresSpoofedForwardedForFromUntrustedPeer(t *testing.T) {
	// Arrange: no trusted proxies, so nothing may set the header
	spoofed := captureDeviceRequest(t, nil, map[string]string{
		"X-Forwarded-For": "203.0.113.7",
		"X-Real-IP":       "198.51.100.2",
	})
	plain := captureDeviceRequest(t, nil, nil)

	// Assert
	assert.NotEqual(t, "203.0.113.7", spoofed.ClientIP, "a spoofed X-Forwarded-For must be ignored")
	assert.NotEqual(t, "198.51.100.2", spoofed.ClientIP, "X-Real-IP is not consulted either")
	assert.Equal(t, plain.ClientIP, spoofed.ClientIP, "sending the header changes nothing")
}

// Behind a real proxy the header is the only way to see the client, so a trusted peer is
// believed. httptest requests arrive from 0.0.0.0, which is what the CIDR below covers.
func TestNewDeviceRequest_HonoursForwardedForFromTrustedProxy(t *testing.T) {
	got := captureDeviceRequest(t, []string{"0.0.0.0/32"}, map[string]string{
		"X-Forwarded-For": "203.0.113.7",
	})

	assert.Equal(t, "203.0.113.7", got.ClientIP)
}
