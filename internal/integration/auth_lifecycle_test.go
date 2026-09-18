package integration

import (
	"net/http"
	"testing"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "an-integration-test-password"

// forEachCacheMode runs one scenario under every auth.session_cache setting.
//
// The scenarios below are all about revocation taking effect, and the cache is the only place
// where the database and the answer a client gets can disagree. Running them once would leave
// the mode that was picked as the only one known to work.
func forEachCacheMode(t *testing.T, scenario func(t *testing.T, s *stack)) {
	t.Helper()

	for _, mode := range cacheModes() {
		t.Run(mode.name, func(t *testing.T) {
			scenario(t, newStack(t, mode))
		})
	}
}

// The core lifecycle. The last step is the point: rotation has to make the old refresh token
// useless, or a stolen one stays valid for the life of the session.
func TestLifecycle_LoginRefreshLogout(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		first := s.login(t, email, testPassword)

		status, _ := s.do(t, http.MethodGet, "/me", first.Access, nil)
		require.Equal(t, http.StatusOK, status, "the access token must work right after login")

		status, second := s.refresh(t, first.Refresh)
		require.Equal(t, http.StatusOK, status)
		assert.NotEqual(t, first.Refresh, second.Refresh, "refresh must rotate the token")
		assert.Equal(t, first.Session, second.Session, "refreshing stays inside the same session")

		status, _ = s.do(t, http.MethodPost, "/logout", second.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodGet, "/me", second.Access, nil)
		assert.Equal(t, http.StatusUnauthorized, status, "the access token dies with the session")

		status, _ = s.refresh(t, second.Refresh)
		assert.Equal(t, http.StatusUnauthorized, status, "the refresh token dies with the session")

		isActive, reason := s.sessionRow(t, first.Session)
		assert.False(t, isActive)
		assert.Equal(t, auth.RevokedReasonLogout, reason)
	})
}

// A rotated refresh token must stop working immediately, not at the end of the grace window.
// The grace only covers the token the last rotation replaced, and only briefly.
func TestLifecycle_TheSupersededRefreshTokenStopsWorking(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		first := s.login(t, email, testPassword)

		status, second := s.refresh(t, first.Refresh)
		require.Equal(t, http.StatusOK, status)

		status, third := s.refresh(t, second.Refresh)
		require.Equal(t, http.StatusOK, status)

		// first.Refresh is now two rotations old, so the grace cannot explain it. That is
		// reuse, and reuse takes the whole session down.
		status, _ = s.refresh(t, first.Refresh)
		assert.Equal(t, http.StatusUnauthorized, status)

		status, _ = s.refresh(t, third.Refresh)
		assert.Equal(t, http.StatusUnauthorized, status,
			"the session was revoked, so the token the attacker raced is dead too")

		isActive, reason := s.sessionRow(t, first.Session)
		assert.False(t, isActive)
		assert.Equal(t, auth.RevokedReasonReuseDetected, reason)
	})
}

// Logging out of one device must not sign the user out everywhere. This is the property a
// single shared "revoke the user's tokens" implementation silently breaks.
func TestLifecycle_LogoutOnOneDeviceLeavesTheOtherSignedIn(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		deviceA := s.login(t, email, testPassword)
		deviceB := s.login(t, email, testPassword)
		require.NotEqual(t, deviceA.Session, deviceB.Session, "two logins are two sessions")

		status, _ := s.do(t, http.MethodPost, "/logout", deviceA.Access, nil)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.do(t, http.MethodGet, "/me", deviceA.Access, nil)
		assert.Equal(t, http.StatusUnauthorized, status)

		status, _ = s.do(t, http.MethodGet, "/me", deviceB.Access, nil)
		assert.Equal(t, http.StatusOK, status, "B was never logged out")

		status, _ = s.refresh(t, deviceB.Refresh)
		assert.Equal(t, http.StatusOK, status, "B can still mint new access tokens")
	})
}

// Reuse detection is deliberately narrow: it revokes the login the replayed token belongs to,
// not the account. Widening it would hand any attacker who captures one token the ability to
// sign the victim out of every device they own.
func TestLifecycle_ReuseRevokesOnlyTheAffectedSession(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		deviceA := s.login(t, email, testPassword)
		deviceB := s.login(t, email, testPassword)

		status, rotatedA := s.refresh(t, deviceA.Refresh)
		require.Equal(t, http.StatusOK, status)

		status, _ = s.refresh(t, rotatedA.Refresh)
		require.Equal(t, http.StatusOK, status)

		// Replaying A's original token: two rotations behind, so outside the grace.
		status, _ = s.refresh(t, deviceA.Refresh)
		require.Equal(t, http.StatusUnauthorized, status)

		isActive, reason := s.sessionRow(t, deviceA.Session)
		assert.False(t, isActive)
		assert.Equal(t, auth.RevokedReasonReuseDetected, reason)

		isActive, reason = s.sessionRow(t, deviceB.Session)
		assert.True(t, isActive, "B had nothing to do with the replay")
		assert.Empty(t, reason)

		status, _ = s.do(t, http.MethodGet, "/me", deviceB.Access, nil)
		assert.Equal(t, http.StatusOK, status)
	})
}

// A refresh token replayed within the grace window is a lost response or a second tab, not an
// attack. It must hand back the session's current token so both callers converge, and it must
// not revoke anything.
func TestLifecycle_ReplayWithinTheGraceIsIdempotent(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		first := s.login(t, email, testPassword)

		status, second := s.refresh(t, first.Refresh)
		require.Equal(t, http.StatusOK, status)

		// Immediately replaying the token the rotation just replaced.
		status, retried := s.refresh(t, first.Refresh)
		require.Equal(t, http.StatusOK, status, "a lost response must not lock the client out")
		assert.Equal(t, second.Session, retried.Session)

		isActive, reason := s.sessionRow(t, first.Session)
		assert.True(t, isActive, "nothing was revoked")
		assert.Empty(t, reason)

		status, _ = s.do(t, http.MethodGet, "/me", retried.Access, nil)
		assert.Equal(t, http.StatusOK, status)
	})
}

// LogoutAll is the account-wide switch, and it has to reach every device.
func TestLifecycle_LogoutAllEndsEverySession(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		deviceA := s.login(t, email, testPassword)
		deviceB := s.login(t, email, testPassword)
		deviceC := s.login(t, email, testPassword)

		status, _ := s.do(t, http.MethodPost, "/logout-all", deviceA.Access, nil)
		require.Equal(t, http.StatusOK, status)

		for name, device := range map[string]tokens{"A": deviceA, "B": deviceB, "C": deviceC} {
			status, _ := s.do(t, http.MethodGet, "/me", device.Access, nil)
			assert.Equal(t, http.StatusUnauthorized, status, "device %s still authenticates", name)

			status, _ = s.refresh(t, device.Refresh)
			assert.Equal(t, http.StatusUnauthorized, status, "device %s can still refresh", name)

			isActive, reason := s.sessionRow(t, device.Session)
			assert.False(t, isActive, "device %s session is still active", name)
			assert.Equal(t, auth.RevokedReasonLogoutAll, reason)
		}
	})
}

// A refresh after a logout is a client that has not noticed yet, not an attack. Recording it
// as reuse would fill the audit trail with false alarms and make the real signal useless.
//
// Two independent guards enforce this — the refresh middleware's session check and
// resolveFailedRotation's own validity check — and either one alone is enough, so this test
// only fails when both are gone. That is the intended defence in depth, not redundancy to be
// cleaned up: removing either leaves the property resting on a single check.
func TestLifecycle_RefreshAfterLogoutIsAPlain401WithNoReuseEvent(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		issued := s.login(t, email, testPassword)

		status, _ := s.do(t, http.MethodPost, "/logout", issued.Access, nil)
		require.Equal(t, http.StatusOK, status)

		events := captureSecurityEvents(t, func() {
			status, _ = s.refresh(t, issued.Refresh)
		})

		assert.Equal(t, http.StatusUnauthorized, status)
		assert.NotContains(t, actionsOf(events), logger.ActionTokenReuseDetected,
			"a logged-out client refreshing is not token reuse")

		// The revocation is still the logout's, not a second one written by the refresh.
		_, reason := s.sessionRow(t, issued.Session)
		assert.Equal(t, auth.RevokedReasonLogout, reason)
	})
}

// Real reuse must be recorded, or nothing distinguishes it from an ordinary rejected refresh.
// The counterpart to the test above: together they pin both sides of the distinction.
func TestLifecycle_ReuseIsRecordedAsASecurityEvent(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		issued := s.login(t, email, testPassword)

		status, rotated := s.refresh(t, issued.Refresh)
		require.Equal(t, http.StatusOK, status)
		status, _ = s.refresh(t, rotated.Refresh)
		require.Equal(t, http.StatusOK, status)

		events := captureSecurityEvents(t, func() {
			status, _ = s.refresh(t, issued.Refresh)
		})

		require.Equal(t, http.StatusUnauthorized, status)

		var reuse map[string]any
		for _, event := range events {
			if event["action"] == logger.ActionTokenReuseDetected {
				require.Nil(t, reuse, "reuse must be recorded exactly once")
				reuse = event
			}
		}

		require.NotNil(t, reuse, "reuse produced no security event: %v", actionsOf(events))
		assert.Equal(t, logger.OutcomeFailure, reuse["outcome"])
		assert.Equal(t, issued.Session, reuse["session_id"])
	})
}

// The audit trail is written to the same stream as everything else, so it is the one place a
// token or password could leak into logs that are shipped off the host and kept for months.
func TestLifecycle_SecurityEventsCarryNoSecrets(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		var issued tokens
		events := captureSecurityEvents(t, func() {
			issued = s.login(t, email, testPassword)
			_, _ = s.do(t, http.MethodPost, "/logout", issued.Access, nil)
		})

		require.NotEmpty(t, events)

		for _, event := range events {
			for key, value := range event {
				text, ok := value.(string)
				if !ok {
					continue
				}
				assert.NotContains(t, text, testPassword, "%s leaked the password", key)
				assert.NotContains(t, text, issued.Access, "%s leaked the access token", key)
				assert.NotContains(t, text, issued.Refresh, "%s leaked the refresh token", key)
			}
		}
	})
}

// An access token issued for a session revoked on another instance must stop working here too.
// With a cache in front of the session store, that only holds if the revocation evicts it —
// which is why this runs under every mode rather than the one the developer happened to use.
func TestLifecycle_RevocationElsewhereIsHonoured(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)
		issued := s.login(t, email, testPassword)

		// Warm the cache: the session is now held wherever this mode holds it.
		status, _ := s.do(t, http.MethodGet, "/me", issued.Access, nil)
		require.Equal(t, http.StatusOK, status)

		// Revoked through the use case, as another instance of the application would.
		require.NoError(t, s.uc.Logout(t.Context(), issued.UserID, issued.Session))

		status, _ = s.do(t, http.MethodGet, "/me", issued.Access, nil)
		assert.Equal(t, http.StatusUnauthorized, status, "a cached session outlived its revocation")
	})
}

// Once an account is locked, the right password and a wrong one must get the same answer.
//
// This is the property the lockout exists for: if the response differed, an attacker could keep
// guessing straight through the lockout and read off which guess was correct, and the lock
// would only be slowing them down rather than stopping them.
//
// Note what is deliberately *not* asserted here: the locked response says the account is locked
// rather than imitating a wrong password. That discloses the account exists — but only to
// someone who just spent the attempts to lock it, and who therefore already knew. Hiding it
// would leave a locked-out user with a correct password and no explanation.
func TestLifecycle_LockedAccountAnswersTheRightPasswordLikeAWrongOne(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		for range testMaxAttempts {
			status, _ := s.do(t, http.MethodPost, "/login", "",
				loginRequest{Email: email, Password: "not-the-password"})
			require.Equal(t, http.StatusUnauthorized, status)
		}

		lockedStatus, lockedBody := s.do(t, http.MethodPost, "/login", "",
			loginRequest{Email: email, Password: testPassword})
		wrongStatus, wrongBody := s.do(t, http.MethodPost, "/login", "",
			loginRequest{Email: email, Password: "still-not-the-password"})

		assert.Equal(t, wrongStatus, lockedStatus)
		assert.Equal(t, string(wrongBody), string(lockedBody),
			"the correct password must not be distinguishable while the account is locked")

		// And the correct password really is refused, not merely answered identically.
		assert.Equal(t, http.StatusUnauthorized, lockedStatus)
	})
}

// Registration and login must not confirm whether an email exists. An unknown address has to
// fail exactly like a known one with the wrong password.
func TestLifecycle_UnknownEmailIsIndistinguishableFromAWrongPassword(t *testing.T) {
	forEachCacheMode(t, func(t *testing.T, s *stack) {
		email := s.createUser(t, testPassword)

		unknownStatus, unknownBody := s.do(t, http.MethodPost, "/login", "",
			loginRequest{Email: "nobody-" + email, Password: testPassword})
		wrongStatus, wrongBody := s.do(t, http.MethodPost, "/login", "",
			loginRequest{Email: email, Password: "not-the-password"})

		assert.Equal(t, wrongStatus, unknownStatus)
		assert.Equal(t, string(wrongBody), string(unknownBody))
	})
}
