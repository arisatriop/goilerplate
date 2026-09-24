package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/constants"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	callerUserID    = "0190a6f0-0000-7000-8000-00000000000a"
	callerSessionID = "0190a6f0-0000-7000-8000-0000000000a1"
	otherSessionID  = "0190a6f0-0000-7000-8000-0000000000a2"
)

// sessionsUsecase records the session-management calls; any other Usecase method panics.
type sessionsUsecase struct {
	auth.Usecase
	sessions  []auth.UserSession
	err       error
	gotUserID string
	gotID     string
	gotKeep   *string
}

func (u *sessionsUsecase) ListSessions(_ context.Context, userID string) ([]auth.UserSession, error) {
	u.gotUserID = userID
	return u.sessions, u.err
}

func (u *sessionsUsecase) RevokeSession(_ context.Context, userID, sessionID string) error {
	u.gotUserID, u.gotID = userID, sessionID
	return u.err
}

func (u *sessionsUsecase) LogoutAll(_ context.Context, userID, keepSessionID string) error {
	u.gotUserID, u.gotKeep = userID, &keepSessionID
	return u.err
}

// newSessionsApp mounts the handlers behind a stand-in for the auth middleware, which is what
// puts the caller's user and session IDs in Locals.
func newSessionsApp(usecase auth.Usecase) *fiber.App {
	h := handler.NewAuth(nil, validator.New(), nil, usecase)

	app := fiber.New()
	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals(string(constants.ContextKeyUserID), callerUserID)
		ctx.Locals(string(constants.ContextKeySessionID), callerSessionID)
		return ctx.Next()
	})
	app.Get("/users/me/sessions", h.ListSessions)
	app.Delete("/users/me/sessions/:id", h.RevokeSession)
	app.Post("/auth/logout-all", h.LogoutAll)
	return app
}

func send(t *testing.T, app *fiber.App, method, path string) (int, map[string]any) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(method, path, nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body), "body: %s", raw)
	return resp.StatusCode, body
}

func TestAuthListSessions_MarksTheCurrentSession(t *testing.T) {
	// Arrange
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	usecase := &sessionsUsecase{sessions: []auth.UserSession{
		{ID: callerSessionID, DeviceName: "Chrome on macOS", DeviceType: auth.DeviceTypeWeb, IPAddress: "203.0.113.7",
			RefreshJTI: "secret-jti", DeviceID: "fp_secret", LastUsedAt: now, CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
		{ID: otherSessionID, DeviceName: "iPhone", DeviceType: auth.DeviceTypeMobile},
	}}

	// Act
	status, body := send(t, newSessionsApp(usecase), http.MethodGet, "/users/me/sessions")

	// Assert
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, callerUserID, usecase.gotUserID, "the user ID must come from the token")

	data := body["data"].([]any)
	require.Len(t, data, 2)
	first := data[0].(map[string]any)
	assert.Equal(t, callerSessionID, first["id"])
	assert.Equal(t, true, first["current"])
	assert.Equal(t, "Chrome on macOS", first["deviceName"])
	assert.Equal(t, "web", first["deviceType"])
	assert.Equal(t, "203.0.113.7", first["ipAddress"])
	assert.Equal(t, "2026-09-24T10:00:00Z", first["lastUsedAt"])
	assert.Equal(t, false, data[1].(map[string]any)["current"])

	// Nothing that could help replay or fingerprint a session is exposed.
	for _, key := range []string{"refreshJti", "deviceId", "userId", "isActive"} {
		assert.NotContains(t, first, key)
	}
}

func TestAuthListSessions_EmptyIsAnArray(t *testing.T) {
	status, body := send(t, newSessionsApp(&sessionsUsecase{}), http.MethodGet, "/users/me/sessions")

	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []any{}, body["data"])
}

func TestAuthRevokeSession(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		err        error
		wantStatus int
		wantCalled bool
	}{
		{"revoked", otherSessionID, nil, http.StatusOK, true},
		{"not found, not yours, or already revoked", otherSessionID,
			auth.ErrSessionNotFound, http.StatusNotFound, true},
		{"malformed id never reaches the database", "not-a-uuid", nil, http.StatusBadRequest, false},
		{"database failure", otherSessionID, context.DeadlineExceeded, http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &sessionsUsecase{err: tt.err}

			status, body := send(t, newSessionsApp(usecase), http.MethodDelete, "/users/me/sessions/"+tt.id)

			assert.Equal(t, tt.wantStatus, status, "body: %v", body)
			if !tt.wantCalled {
				assert.Empty(t, usecase.gotID)
				return
			}
			assert.Equal(t, callerUserID, usecase.gotUserID)
			assert.Equal(t, tt.id, usecase.gotID)
		})
	}
}

func TestAuthLogoutAll_KeepCurrent(t *testing.T) {
	tests := []struct {
		query    string
		wantKeep string
	}{
		{"", ""},
		{"?keep_current=false", ""},
		{"?keep_current=true", callerSessionID},
	}

	for _, tt := range tests {
		t.Run("query="+tt.query, func(t *testing.T) {
			usecase := &sessionsUsecase{}

			status, _ := send(t, newSessionsApp(usecase), http.MethodPost, "/auth/logout-all"+tt.query)

			require.Equal(t, http.StatusOK, status)
			require.NotNil(t, usecase.gotKeep)
			assert.Equal(t, tt.wantKeep, *usecase.gotKeep)
		})
	}
}

func TestAuthLogoutAll_MalformedQueryIs400(t *testing.T) {
	usecase := &sessionsUsecase{}

	status, _ := send(t, newSessionsApp(usecase), http.MethodPost, "/auth/logout-all?keep_current=maybe")

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, usecase.gotKeep, "nothing is revoked on a request that was refused")
}
