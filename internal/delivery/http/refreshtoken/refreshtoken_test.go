package refreshtoken_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goilerplate/internal/delivery/http/refreshtoken"
	"goilerplate/pkg/apperr"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issuedCookie runs Issue in a request and returns the Set-Cookie the client received.
func issuedCookie(t *testing.T, transport *refreshtoken.Transport) (*http.Cookie, string) {
	t.Helper()

	var body string
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		body = transport.Issue(c, "the-refresh-token", time.Now().Add(time.Hour))
		return c.SendStatus(fiber.StatusNoContent)
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return nil, body
	}
	return cookies[0], body
}

func TestIssue_CookieModeKeepsTheTokenOutOfTheBody(t *testing.T) {
	transport := refreshtoken.New(refreshtoken.Options{Cookie: true, Secure: true})

	cookie, body := issuedCookie(t, transport)

	assert.Empty(t, body, "in cookie mode the token must not reach the JSON body")
	require.NotNil(t, cookie)
	assert.Equal(t, "__Secure-refresh_token", cookie.Name, "the prefix makes browsers refuse it over plain HTTP")
	assert.Equal(t, "the-refresh-token", cookie.Value)
	assert.True(t, cookie.HttpOnly, "page JavaScript must not be able to read it")
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, refreshtoken.CookiePath, cookie.Path, "only the auth endpoints receive it")
	assert.False(t, cookie.Expires.IsZero(), "it expires with the refresh token")
}

func TestIssue_LocalCookieIsNotSecureAndDropsThePrefix(t *testing.T) {
	cookie, _ := issuedCookie(t, refreshtoken.New(refreshtoken.Options{Cookie: true, SameSite: "lax"}))

	require.NotNil(t, cookie)
	assert.Equal(t, "refresh_token", cookie.Name)
	assert.False(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestIssue_BodyModeSetsNoCookie(t *testing.T) {
	for name, transport := range map[string]*refreshtoken.Transport{
		"explicit": refreshtoken.New(refreshtoken.Options{}),
		"nil":      nil,
	} {
		t.Run(name, func(t *testing.T) {
			cookie, body := issuedCookie(t, transport)

			assert.Nil(t, cookie)
			assert.Equal(t, "the-refresh-token", body)
		})
	}
}

func read(t *testing.T, transport *refreshtoken.Transport, headers map[string]string) (string, error) {
	t.Helper()

	var token string
	var readErr error
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		token, readErr = transport.Read(c)
		return c.SendStatus(fiber.StatusNoContent)
	})
	req := httptest.NewRequest(fiber.MethodPost, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	_, err := app.Test(req)
	require.NoError(t, err)
	return token, readErr
}

// Each mode accepts only its own transport: a deployment has one place the refresh token lives.
func TestRead_EachModeAcceptsOnlyItsOwnTransport(t *testing.T) {
	cookieMode := refreshtoken.New(refreshtoken.Options{Cookie: true, Secure: true})

	token, err := read(t, cookieMode, map[string]string{"Cookie": "__Secure-refresh_token=from-cookie"})
	require.NoError(t, err)
	assert.Equal(t, "from-cookie", token)

	_, err = read(t, cookieMode, map[string]string{"Authorization": "Bearer from-header"})
	assert.Error(t, err, "cookie mode must not fall back to the header")

	token, err = read(t, nil, map[string]string{"Authorization": "Bearer from-header"})
	require.NoError(t, err)
	assert.Equal(t, "from-header", token)

	_, err = read(t, nil, map[string]string{"Cookie": "refresh_token=from-cookie"})
	assert.Error(t, err, "body mode must not read a cookie")
}

func TestBearerToken_MalformedHeadersAreOneAnswer(t *testing.T) {
	for _, header := range []string{"", "Bearer", "Bearer   ", "Basic abc", "bearerabc"} {
		_, err := read(t, nil, map[string]string{"Authorization": header})

		appErr, ok := apperr.As(err)
		require.True(t, ok, "header %q", header)
		assert.Equal(t, "unauthorized", appErr.Code)
	}
}

func checkOrigin(t *testing.T, transport *refreshtoken.Transport, origin string) error {
	t.Helper()

	var checkErr error
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		checkErr = transport.CheckOrigin(c)
		return c.SendStatus(fiber.StatusNoContent)
	})
	req := httptest.NewRequest(fiber.MethodPost, "http://api.example.com/", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	_, err := app.Test(req)
	require.NoError(t, err)
	return checkErr
}

// Browsers always send Origin on a cross-site POST, so a present Origin must be ours or listed.
// An absent one is not a browser page and cannot be CSRF — curl, a mobile app, a server.
func TestCheckOrigin(t *testing.T) {
	transport := refreshtoken.New(refreshtoken.Options{
		Cookie:         true,
		AllowedOrigins: []string{"https://app.example.com", "https://*.example.net", "*"},
	})

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"", true},
		{"http://api.example.com", true},
		{"https://app.example.com", true},
		{"https://shop.example.net", true},
		{"https://evil.example.org", false},
		{"https://example.net", false},
		{"http://shop.example.net", false},
		{"https://app.example.com.evil.org", false},
		{"null", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			err := checkOrigin(t, transport, tt.origin)
			if tt.allowed {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, refreshtoken.ErrOriginNotAllowed)
		})
	}
}

func TestCheckOrigin_BodyModeDoesNotCheck(t *testing.T) {
	assert.NoError(t, checkOrigin(t, nil, "https://evil.example.org"),
		"a Bearer header is never attached by the browser, so there is nothing to forge")
}

func TestClear_ExpiresTheCookie(t *testing.T) {
	transport := refreshtoken.New(refreshtoken.Options{Cookie: true, Secure: true})
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		transport.Clear(c)
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/", nil))
	require.NoError(t, err)

	require.Len(t, resp.Cookies(), 1)
	cleared := resp.Cookies()[0]
	assert.Equal(t, "__Secure-refresh_token", cleared.Name)
	assert.Empty(t, cleared.Value)
	assert.True(t, cleared.Expires.Before(time.Now()), "an expired cookie is how a server deletes one")
	assert.Equal(t, refreshtoken.CookiePath, cleared.Path, "must match the path it was set on, or the browser keeps the original")
}
