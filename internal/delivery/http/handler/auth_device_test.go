package handler

import (
	"net/http/httptest"
	"testing"

	"goilerplate/internal/domain/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeviceRequest_CopiesRequestAttributes(t *testing.T) {
	// Arrange
	var got auth.DeviceRequest
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		got = newDeviceRequest(ctx)
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Accept-Language", "id-ID")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	req.Header.Set("X-Real-IP", "198.51.100.2")

	// Act
	_, err := app.Test(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-agent", got.UserAgent)
	assert.Equal(t, "id-ID", got.AcceptLanguage)
	assert.Equal(t, "gzip", got.AcceptEncoding)
	assert.Equal(t, "203.0.113.7", got.ForwardedFor)
	assert.Equal(t, "198.51.100.2", got.RealIP)
	assert.NotEmpty(t, got.RemoteIP)
}
