package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"goilerplate/internal/domain/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listSessions returns the IDs of the caller's active sessions, in the order the API gives them.
func (s *stack) listSessions(t *testing.T, access string) []string {
	t.Helper()

	status, body := s.do(t, http.MethodGet, "/sessions", access, nil)
	require.Equal(t, http.StatusOK, status, "list sessions: %s", body)

	var ids []string
	require.NoError(t, json.Unmarshal(body, &ids))
	return ids
}

// Signing out one other device is the point of the feature: that device loses access at once,
// and the one doing the signing out keeps it.
func TestSessions_RevokingAnotherDeviceSignsOnlyThatDeviceOut(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		phone := s.login(t, email, testPassword)
		laptop := s.login(t, email, testPassword)

		// Read through the cache first, so a stale entry would be there to be served.
		status, _ := s.do(t, http.MethodGet, "/me", laptop.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodDelete, "/sessions/"+laptop.Session, phone.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodGet, "/me", laptop.Access, nil)
		assert.Equal(t, http.StatusUnauthorized, status, "the revoked device still authenticates")
		status, _ = s.refresh(t, laptop.Refresh)
		assert.Equal(t, http.StatusUnauthorized, status, "the revoked device can still refresh")

		status, _ = s.do(t, http.MethodGet, "/me", phone.Access, nil)
		assert.Equal(t, http.StatusOK, status, "the device that did the revoking was signed out too")

		assert.Equal(t, []string{phone.Session}, s.listSessions(t, phone.Access))

		isActive, reason := s.sessionRow(t, laptop.Session)
		assert.False(t, isActive)
		assert.Equal(t, auth.RevokedReasonLogout, reason)
	})
}

// Ownership: the session ID is in the URL, so it is attacker-controlled. Someone else's session
// must answer exactly like one that does not exist, and must survive the attempt.
func TestSessions_SomeoneElsesSessionIsNotFoundAndSurvives(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		attacker := s.login(t, s.createUser(t, testPassword), testPassword)
		victim := s.login(t, s.createUser(t, testPassword), testPassword)

		status, foreign := s.do(t, http.MethodDelete, "/sessions/"+victim.Session, attacker.Access, nil)
		assert.Equal(t, http.StatusNotFound, status)

		_, missing := s.do(t, http.MethodDelete, "/sessions/"+"0190a6f0-0000-7000-8000-000000000000", attacker.Access, nil)
		assert.JSONEq(t, string(missing), string(foreign),
			"someone else's session must be indistinguishable from a missing one")

		status, _ = s.do(t, http.MethodGet, "/me", victim.Access, nil)
		assert.Equal(t, http.StatusOK, status, "the victim was signed out by another user")
		isActive, _ := s.sessionRow(t, victim.Session)
		assert.True(t, isActive)
	})
}

func TestSessions_RevokingTwiceIsNotFound(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		phone := s.login(t, email, testPassword)
		laptop := s.login(t, email, testPassword)

		status, _ := s.do(t, http.MethodDelete, "/sessions/"+laptop.Session, phone.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodDelete, "/sessions/"+laptop.Session, phone.Access, nil)
		assert.Equal(t, http.StatusNotFound, status)
	})
}

// "Sign out everywhere else": the caller keeps working, every other device is turned out.
func TestSessions_LogoutAllCanKeepTheCurrentSession(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		current := s.login(t, email, testPassword)
		otherA := s.login(t, email, testPassword)
		otherB := s.login(t, email, testPassword)

		status, _ := s.do(t, http.MethodPost, "/logout-all?keep_current=true", current.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodGet, "/me", current.Access, nil)
		assert.Equal(t, http.StatusOK, status, "the kept session was signed out")
		status, _ = s.refresh(t, current.Refresh)
		assert.Equal(t, http.StatusOK, status, "the kept session can no longer refresh")

		for name, device := range map[string]tokens{"A": otherA, "B": otherB} {
			status, _ := s.do(t, http.MethodGet, "/me", device.Access, nil)
			assert.Equal(t, http.StatusUnauthorized, status, "device %s still authenticates", name)

			isActive, reason := s.sessionRow(t, device.Session)
			assert.False(t, isActive, "device %s session is still active", name)
			assert.Equal(t, auth.RevokedReasonLogoutAll, reason)
		}
	})
}

func TestSessions_ListShowsOnlyTheCallersActiveSessions(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		first := s.login(t, email, testPassword)
		second := s.login(t, email, testPassword)
		loggedOut := s.login(t, email, testPassword)
		s.login(t, s.createUser(t, testPassword), testPassword) // someone else

		status, _ := s.do(t, http.MethodPost, "/logout", loggedOut.Access, nil)
		require.Equal(t, http.StatusOK, status)

		assert.ElementsMatch(t, []string{first.Session, second.Session}, s.listSessions(t, first.Access))
	})
}
