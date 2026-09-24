package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/config"
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/delivery/http/refreshtoken"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiOrigin = "http://api.example.com"

// cookieApp mounts the real auth handler and middleware — the same routes and order as
// router/public.go — in cookie mode. Unlike the rest of this suite it goes through the handlers,
// because setting, reading and clearing the cookie is exactly what they are responsible for.
func (s *stack) cookieApp() *fiber.App {
	transport := refreshtoken.New(refreshtoken.Options{
		Cookie:         true,
		Secure:         true,
		AllowedOrigins: []string{"https://app.example.com"},
	})
	mw := middleware.NewAuth(s.jwt, s.repo, s.sessions, s.permissions, nil, config.InternalAuth{}, transport)
	h := handler.NewAuth(auth.NewDeviceService(), response.NewValidator(), nil, s.uc, transport)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	group := app.Group("/api/v1/auth")
	group.Post("/login", h.Login)
	group.Post("/refresh", mw.AuthenticateRefreshToken(), h.RefreshToken)
	group.Post("/logout", mw.Authenticate(), h.Logout)
	app.Get("/api/v1/me", mw.Authenticate(), func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	return app
}

type cookieResponse struct {
	status int
	body   map[string]any
	cookie *http.Cookie // the refresh cookie the response set, if any
}

func cookieCall(t *testing.T, app *fiber.App, method, path string, headers map[string]string, body string) cookieResponse {
	t.Helper()

	req := httptest.NewRequest(method, apiOrigin+path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	out := cookieResponse{status: resp.StatusCode}
	if len(raw) > 0 {
		require.NoError(t, json.Unmarshal(raw, &out.body), "body: %s", raw)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "__Secure-refresh_token" {
			out.cookie = c
		}
	}
	return out
}

func withCookie(value string, extra ...string) map[string]string {
	headers := map[string]string{"Cookie": "__Secure-refresh_token=" + value}
	for i := 0; i+1 < len(extra); i += 2 {
		headers[extra[i]] = extra[i+1]
	}
	return headers
}

func cookieLogin(t *testing.T, app *fiber.App, email string) (access string, refreshCookie *http.Cookie) {
	t.Helper()

	res := cookieCall(t, app, http.MethodPost, "/api/v1/auth/login", nil,
		`{"email":"`+email+`","password":"`+testPassword+`"}`)
	require.Equal(t, http.StatusOK, res.status, "login: %v", res.body)
	require.NotNil(t, res.cookie, "login must set the refresh cookie")

	tokens := res.body["data"].(map[string]any)["tokens"].(map[string]any)
	return tokens["accessToken"].(string), res.cookie
}

// The point of cookie mode: page JavaScript never sees the refresh token, and refresh still works.
func TestRefreshCookie_LoginSetsAnHttpOnlyCookieAndNoBodyToken(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		app := s.cookieApp()
		access, cookie := cookieLogin(t, app, s.createUser(t, testPassword))

		assert.NotEmpty(t, access, "the access token still comes in the body")
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		assert.Equal(t, refreshtoken.CookiePath, cookie.Path)

		res := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", withCookie(cookie.Value), "")
		require.Equal(t, http.StatusOK, res.status, "refresh: %v", res.body)
		tokens := res.body["data"].(map[string]any)["tokens"].(map[string]any)
		assert.NotContains(t, tokens, "refreshToken", "the rotated token must not reach the body either")
		require.NotNil(t, res.cookie, "refresh must rotate the cookie")
		assert.NotEqual(t, cookie.Value, res.cookie.Value)
	})
}

// Rotation and reuse detection behave exactly as they do in body mode.
func TestRefreshCookie_ReplayingAnOldCookieRevokesTheSession(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		app := s.cookieApp()
		access, first := cookieLogin(t, app, s.createUser(t, testPassword))

		second := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", withCookie(first.Value), "")
		require.Equal(t, http.StatusOK, second.status)
		third := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", withCookie(second.cookie.Value), "")
		require.Equal(t, http.StatusOK, third.status)

		// Two rotations behind, so outside the grace window: a stolen cookie being replayed.
		replay := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", withCookie(first.Value), "")
		assert.Equal(t, http.StatusUnauthorized, replay.status)

		res := cookieCall(t, app, http.MethodGet, "/api/v1/me", map[string]string{"Authorization": "Bearer " + access}, "")
		assert.Equal(t, http.StatusUnauthorized, res.status, "reuse must end the whole session")
	})
}

// A page on another origin can make the browser send the cookie; it must not get a token for it.
// The refused attempt must not rotate anything either, or the attacker could log the user out.
func TestRefreshCookie_ForeignOriginIsRefusedAndRotatesNothing(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		app := s.cookieApp()
		_, cookie := cookieLogin(t, app, s.createUser(t, testPassword))

		forged := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh",
			withCookie(cookie.Value, "Origin", "https://evil.example.org"), "")
		assert.Equal(t, http.StatusForbidden, forged.status)
		assert.Equal(t, "origin_not_allowed", forged.body["code"])
		assert.Nil(t, forged.cookie)

		for _, origin := range []string{"https://app.example.com", apiOrigin, ""} {
			headers := withCookie(cookie.Value)
			if origin != "" {
				headers["Origin"] = origin
			}
			res := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", headers, "")
			require.Equal(t, http.StatusOK, res.status, "origin %q: %v", origin, res.body)
			cookie = res.cookie
		}
	})
}

// Cookie mode reads only the cookie, so a refresh token lifted from somewhere and sent as a
// header gets nothing.
func TestRefreshCookie_BearerHeaderIsNotAccepted(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		app := s.cookieApp()
		_, cookie := cookieLogin(t, app, s.createUser(t, testPassword))

		res := cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh",
			map[string]string{"Authorization": "Bearer " + cookie.Value}, "")

		assert.Equal(t, http.StatusUnauthorized, res.status)
	})
}

func TestRefreshCookie_LogoutClearsTheCookieAndEndsTheSession(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		app := s.cookieApp()
		access, cookie := cookieLogin(t, app, s.createUser(t, testPassword))

		res := cookieCall(t, app, http.MethodPost, "/api/v1/auth/logout", map[string]string{"Authorization": "Bearer " + access}, "")
		require.Equal(t, http.StatusOK, res.status)
		require.NotNil(t, res.cookie, "logout must clear the cookie")
		assert.Empty(t, res.cookie.Value)
		assert.Equal(t, refreshtoken.CookiePath, res.cookie.Path)

		res = cookieCall(t, app, http.MethodPost, "/api/v1/auth/refresh", withCookie(cookie.Value), "")
		assert.Equal(t, http.StatusUnauthorized, res.status, "a cookie kept past logout must not work")
	})
}
