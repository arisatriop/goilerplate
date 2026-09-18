package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/redact"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPassword     = "hunter2-password"
	testAccessToken  = "eyJaccess.secret.token"
	testRefreshToken = "eyJrefresh.secret.token"
	testAPIKey       = "partner-api-key-123"
	testResetToken   = "reset-token-xyz"
)

// captureLogs routes the default slog logger into a buffer for the duration of the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	t.Setenv("APP_ENV", "test")

	buf := &bytes.Buffer{}
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	return buf
}

func newLoggedApp() *fiber.App {
	app := fiber.New()
	app.Use(NewRequestLogger(nil).LogRequest())

	tokens := func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"message": "ok",
			"data": fiber.Map{
				"access_token":  testAccessToken,
				"refresh_token": testRefreshToken,
			},
		})
	}
	app.Post("/api/v1/auth/login", tokens)
	app.Post("/api/v1/auth/refresh", tokens)
	app.Post("/api/v1/auth/logout", tokens)
	app.Post("/api/v1/users/me/password", tokens)
	app.Get("/partner/v1/orders", tokens)

	return app
}

func doRequest(t *testing.T, app *fiber.App, method, target, body string, headers map[string]string) {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func TestRequestLogger_LogRequest_NoSecretsInLogs(t *testing.T) {
	// Arrange
	logs := captureLogs(t)
	app := newLoggedApp()
	jsonHeaders := map[string]string{"Content-Type": "application/json"}
	bearer := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + testAccessToken,
		"Cookie":        "refresh_token=" + testRefreshToken,
	}

	// Act: login → refresh → logout, plus a non-auth route and a partner call
	doRequest(t, app, fiber.MethodPost, "/api/v1/auth/login",
		`{"email":"user@example.com","password":"`+testPassword+`"}`, jsonHeaders)
	doRequest(t, app, fiber.MethodPost, "/API/V1/AUTH/refresh",
		`{"refresh_token":"`+testRefreshToken+`"}`, bearer)
	doRequest(t, app, fiber.MethodPost, "/api/v1/auth/logout", "", bearer)
	doRequest(t, app, fiber.MethodPost, "/api/v1/users/me/password",
		`{"current_password":"`+testPassword+`","new_password":"`+testPassword+`"}`, bearer)
	doRequest(t, app, fiber.MethodGet, "/partner/v1/orders?token="+testResetToken+"&page=1", "",
		map[string]string{"x-api-key": testAPIKey})

	// Assert
	output := logs.String()
	require.Equal(t, 5, strings.Count(output, LogLabel), "every request must be logged")
	for _, secret := range []string{testPassword, testAccessToken, testRefreshToken, testAPIKey, testResetToken} {
		assert.NotContains(t, output, secret)
	}
	assert.Contains(t, output, redact.Mask)
	assert.Contains(t, output, redact.Omitted)
	assert.Contains(t, output, "page=1", "non-sensitive query parameters stay readable")
}

func TestRequestLogger_LogRequest_RedactsNonAuthBody(t *testing.T) {
	// Arrange
	logs := captureLogs(t)
	app := newLoggedApp()

	// Act
	doRequest(t, app, fiber.MethodPost, "/api/v1/users/me/password",
		`{"current_password":"`+testPassword+`","keep":"visible"}`,
		map[string]string{"Content-Type": "application/json"})

	// Assert
	output := logs.String()
	assert.NotContains(t, output, testPassword)
	assert.Contains(t, output, `"keep":"visible"`)
}

func TestRequestLogger_ShouldOmitBody(t *testing.T) {
	rl := NewRequestLogger([]string{"api/v1/auth/", " /internal "})

	assert.True(t, rl.shouldOmitBody("/api/v1/auth"))
	assert.True(t, rl.shouldOmitBody("/api/v1/auth/login"))
	assert.True(t, rl.shouldOmitBody("/Api/V1/Auth/Login"))
	assert.True(t, rl.shouldOmitBody("/internal/users"))
	assert.False(t, rl.shouldOmitBody("/api/v1/authors"))
	assert.False(t, rl.shouldOmitBody("/api/v1/users"))
}

// The caller's identity has to reach the context, because security events are raised in the
// auth domain, which has no access to the HTTP request. Without this the audit trail would
// record every login attempt with an empty client IP.
func TestRequestLogger_PutsCallerOnContext(t *testing.T) {
	var gotIP, gotUserAgent, gotRequestID string

	app := fiber.New()
	app.Use(NewRequestLogger(nil).LogRequest())
	app.Get("/", func(ctx *fiber.Ctx) error {
		userCtx := ctx.UserContext()
		gotIP, _ = userCtx.Value(constants.ContextKeyClientIP).(string)
		gotUserAgent, _ = userCtx.Value(constants.ContextKeyUserAgent).(string)
		gotRequestID, _ = userCtx.Value(constants.ContextKeyRequestID).(string)
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	doRequest(t, app, fiber.MethodGet, "/", "", map[string]string{"User-Agent": "audit-agent"})

	assert.Equal(t, "audit-agent", gotUserAgent)
	assert.NotEmpty(t, gotRequestID)
	assert.NotEmpty(t, gotIP, "the client IP must be resolved before any handler runs")
}

// With no trusted proxies the app trusts nobody, so a client cannot choose the address its
// security events are filed under — the same guarantee T4.6 gives c.IP().
func TestRequestLogger_ContextIPIgnoresSpoofedHeader(t *testing.T) {
	capture := func(headers map[string]string) string {
		var got string
		app := fiber.New()
		app.Use(NewRequestLogger(nil).LogRequest())
		app.Get("/", func(ctx *fiber.Ctx) error {
			got, _ = ctx.UserContext().Value(constants.ContextKeyClientIP).(string)
			return ctx.SendStatus(fiber.StatusNoContent)
		})
		doRequest(t, app, fiber.MethodGet, "/", "", headers)
		return got
	}

	spoofed := capture(map[string]string{"X-Forwarded-For": "203.0.113.7"})
	plain := capture(nil)

	assert.NotEqual(t, "203.0.113.7", spoofed)
	assert.Equal(t, plain, spoofed)
}
