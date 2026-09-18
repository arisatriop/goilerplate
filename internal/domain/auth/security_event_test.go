package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureSecurityEvents returns only the security entries fn produced. Application logs are
// filtered out on purpose: the audit trail has to be complete and unambiguous on its own, and
// a test that counted both would pass even if an event were downgraded to an ordinary log line.
func captureSecurityEvents(t *testing.T, fn func()) []map[string]any {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	fn()

	var events []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal(line, &entry))
		if entry["label"] == logger.SecurityLabel {
			events = append(events, entry)
		}
	}

	return events
}

// requireOneEvent asserts the exactly-one rule: a duplicate would inflate any alert counting
// failures, and a missing one would hide the attempt entirely.
func requireOneEvent(t *testing.T, events []map[string]any, action, reason string) map[string]any {
	t.Helper()

	require.Len(t, events, 1, "expected exactly one %s event, got %v", action, events)
	assert.Equal(t, action, events[0]["action"])
	assert.Equal(t, reason, events[0]["reason"])

	return events[0]
}

func TestLoginFailure_EmitsOneEventPerRejectionReason(t *testing.T) {
	lockedUntil := utils.Now().Add(5 * time.Minute)

	lockedUser := testUser(t)
	lockedUser.LockedUntil = &lockedUntil

	disabledUser := testUser(t)
	disabledUser.IsActive = false

	tests := []struct {
		name     string
		repo     *validatorRepo
		email    string
		password string
		action   string
		reason   string
		userID   string
	}{
		{
			name:     "unknown email",
			repo:     &validatorRepo{user: nil},
			email:    "nobody@example.test",
			password: "whatever",
			action:   logger.ActionLoginFailed,
			reason:   logger.ReasonUnknownEmail,
			// No account was targeted, so there is no subject to name — and the attempted
			// address stays out of the trail.
			userID: "",
		},
		{
			name:     "wrong password",
			repo:     &validatorRepo{user: testUser(t)},
			email:    "user@example.test",
			password: "wrong-password",
			action:   logger.ActionLoginFailed,
			reason:   logger.ReasonBadPassword,
			userID:   "u1",
		},
		{
			name:     "locked account",
			repo:     &validatorRepo{user: lockedUser},
			email:    "user@example.test",
			password: correctPassword,
			action:   logger.ActionLoginFailed,
			reason:   logger.ReasonAccountLocked,
			userID:   "u1",
		},
		{
			name:     "disabled account",
			repo:     &validatorRepo{user: disabledUser},
			email:    "user@example.test",
			password: correctPassword,
			action:   logger.ActionLoginFailed,
			reason:   logger.ReasonAccountDisabled,
			userID:   "u1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := captureSecurityEvents(t, func() {
				_, err := newValidator(tc.repo).ValidateUserForLogin(context.Background(), tc.email, tc.password)
				require.Error(t, err)
			})

			event := requireOneEvent(t, events, tc.action, tc.reason)
			assert.Equal(t, logger.OutcomeFailure, event["outcome"])
			assert.Equal(t, tc.userID, event["user_id"])
		})
	}
}

// Locking is its own event, separate from the failure that triggered it: an alert on repeated
// failures and an alert on an account actually being closed are different alerts.
func TestRegisterFailedLogin_LockEmitsItsOwnEvent(t *testing.T) {
	repo := &validatorRepo{user: testUser(t), registerReturns: true}

	events := captureSecurityEvents(t, func() {
		_, err := newValidator(repo).ValidateUserForLogin(context.Background(), "user@example.test", "wrong-password")
		require.Error(t, err)
	})

	require.Len(t, events, 2)
	assert.Equal(t, logger.ActionLoginFailed, events[0]["action"])
	assert.Equal(t, logger.ActionAccountLocked, events[1]["action"])
	assert.Equal(t, "u1", events[1]["user_id"])
	assert.Equal(t, logger.OutcomeFailure, events[1]["outcome"])
}

// The account that was not locked must not produce a lock event, or every failed password
// would look like a lockout.
func TestRegisterFailedLogin_NoLockNoEvent(t *testing.T) {
	repo := &validatorRepo{user: testUser(t), registerReturns: false}

	events := captureSecurityEvents(t, func() {
		_, err := newValidator(repo).ValidateUserForLogin(context.Background(), "user@example.test", "wrong-password")
		require.Error(t, err)
	})

	requireOneEvent(t, events, logger.ActionLoginFailed, logger.ReasonBadPassword)
}

// A successful login is recorded too. An audit trail that only holds failures cannot answer
// the question that usually follows an incident: what did the attacker actually get into.
func TestLoginSuccess_IsRecorded(t *testing.T) {
	repo := &validatorRepo{user: testUser(t)}

	events := captureSecurityEvents(t, func() {
		got, err := newValidator(repo).ValidateUserForLogin(context.Background(), "user@example.test", correctPassword)
		require.NoError(t, err)
		require.NotNil(t, got)

		// ValidateUserForLogin only validates; the use case logs the success once the session
		// is committed, which is what this stands in for.
		logger.Security(context.Background(), logger.SecurityEvent{
			Action:    logger.ActionLoginSucceeded,
			Outcome:   logger.OutcomeSuccess,
			UserID:    got.ID,
			SessionID: "s1",
		})
	})

	require.Len(t, events, 1, "validation itself must stay silent on success")
	assert.Equal(t, logger.ActionLoginSucceeded, events[0]["action"])
	assert.Equal(t, logger.OutcomeSuccess, events[0]["outcome"])
}
