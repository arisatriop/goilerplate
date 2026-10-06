package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const newTestPassword = "a-different-integration-password"

type resetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// forgot asks for a reset link for email and returns the token from the newest email.
func (s *stack) forgot(t *testing.T, email string) string {
	t.Helper()

	status, body := s.do(t, http.MethodPost, "/forgot-password", "", map[string]string{"email": email})
	require.Equal(t, http.StatusOK, status, "forgot-password: %s", body)

	notices := s.inbox.received(email)
	require.NotEmpty(t, notices, "no reset email reached %s", email)
	return notices[len(notices)-1].Token
}

// The point of a reset: whoever got into the account is turned out. Every device's access and
// refresh tokens stop working at once, and only the new password signs in. Run under every
// cache mode, because a session served from a cache that missed the eviction would survive.
func TestPasswordReset_TurnsOutEverySessionAndTheOldPassword(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		laptop := s.login(t, email, testPassword)
		phone := s.login(t, email, testPassword)

		// Warm the cache, so an eviction that did not happen would show.
		for _, device := range []tokens{laptop, phone} {
			status, _ := s.do(t, http.MethodGet, "/me", device.Access, nil)
			require.Equal(t, http.StatusOK, status)
		}

		token := s.forgot(t, email)
		status, body := s.do(t, http.MethodPost, "/reset-password", "", resetRequest{token, newTestPassword})
		require.Equal(t, http.StatusOK, status, "reset-password: %s", body)

		for name, device := range map[string]tokens{"laptop": laptop, "phone": phone} {
			status, _ := s.do(t, http.MethodGet, "/me", device.Access, nil)
			assert.Equal(t, http.StatusUnauthorized, status, "%s access token survived the reset", name)

			status, _ = s.refresh(t, device.Refresh)
			assert.Equal(t, http.StatusUnauthorized, status, "%s refresh token survived the reset", name)

			isActive, reason := s.sessionRow(t, device.Session)
			assert.False(t, isActive)
			assert.Equal(t, auth.RevokedReasonPasswordReset, reason)
		}

		status, _ = s.do(t, http.MethodPost, "/login", "", loginRequest{Email: email, Password: testPassword})
		assert.Equal(t, http.StatusUnauthorized, status, "the old password stops working")
		s.login(t, email, newTestPassword)
	})
}

// A link works once, and only the newest one works at all.
func TestPasswordReset_LinksAreSingleUseAndSuperseded(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)

	older := s.forgot(t, email)
	newer := s.forgot(t, email)
	require.NotEqual(t, older, newer)

	status, body := s.do(t, http.MethodPost, "/reset-password", "", resetRequest{older, newTestPassword})
	assert.Equal(t, http.StatusBadRequest, status, "a superseded link must not work")
	assertCode(t, body, "invalid_reset_token")

	status, _ = s.do(t, http.MethodPost, "/reset-password", "", resetRequest{newer, newTestPassword})
	require.Equal(t, http.StatusOK, status)

	status, body = s.do(t, http.MethodPost, "/reset-password", "", resetRequest{newer, "yet-another-password"})
	assert.Equal(t, http.StatusBadRequest, status, "a used link must not work again")
	assertCode(t, body, "invalid_reset_token")

	s.login(t, email, newTestPassword)
}

// An unknown address is answered exactly like a registered one, sends nothing, and is visible
// only in the security log.
func TestPasswordReset_UnknownAddressLooksLikeAKnownOne(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	known := s.createUser(t, testPassword)
	const unknown = "nobody-here@example.test"

	var knownStatus, unknownStatus int
	var knownBody, unknownBody []byte
	events := captureSecurityEvents(t, func() {
		knownStatus, knownBody = s.do(t, http.MethodPost, "/forgot-password", "", map[string]string{"email": known})
		unknownStatus, unknownBody = s.do(t, http.MethodPost, "/forgot-password", "", map[string]string{"email": unknown})
	})

	assert.Equal(t, knownStatus, unknownStatus)
	assert.Equal(t, knownBody, unknownBody)
	assert.Len(t, s.inbox.received(known), 1)
	assert.Empty(t, s.inbox.received(unknown))

	require.Len(t, events, 2)
	assert.Equal(t, []string{logger.ActionPasswordResetRequested, logger.ActionPasswordResetRequested}, actionsOf(events))
	assert.Equal(t, logger.OutcomeSuccess, events[0]["outcome"])
	assert.Equal(t, logger.OutcomeFailure, events[1]["outcome"])
	assert.Equal(t, logger.ReasonUnknownEmail, events[1]["reason"])
}

func assertCode(t *testing.T, body []byte, want string) {
	t.Helper()
	var envelope struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	assert.Equal(t, want, envelope.Code)
}
