package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"goilerplate/pkg/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureSecurity swaps the default logger for the duration of one test and returns every
// entry fn produced, decoded. Security() writes through slog.Default(), which is how the
// application is wired, so this exercises the real path rather than a parallel one.
func captureSecurity(t *testing.T, fn func()) []map[string]any {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	fn()

	var entries []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal(line, &entry))
		entries = append(entries, entry)
	}

	return entries
}

// requestContext builds what the request-logger middleware leaves on the context.
func requestContext() context.Context {
	ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "req-1")
	ctx = context.WithValue(ctx, constants.ContextKeyClientIP, "203.0.113.7")
	ctx = context.WithValue(ctx, constants.ContextKeyUserAgent, "test-agent")
	return ctx
}

func TestSecurity_WritesOneLabelledEntry(t *testing.T) {
	entries := captureSecurity(t, func() {
		Security(requestContext(), SecurityEvent{
			Action:    ActionLoginSucceeded,
			Outcome:   OutcomeSuccess,
			UserID:    "u1",
			SessionID: "s1",
		})
	})

	require.Len(t, entries, 1, "one event must produce exactly one entry")
	entry := entries[0]

	// The label is what a SIEM filters on, so it must not be shared with application logs.
	assert.Equal(t, SecurityLabel, entry["label"])
	assert.NotEqual(t, LogLabel, entry["label"])

	assert.Equal(t, ActionLoginSucceeded, entry["action"])
	assert.Equal(t, OutcomeSuccess, entry["outcome"])
	assert.Equal(t, "u1", entry["user_id"])
	assert.Equal(t, "s1", entry["session_id"])
	assert.Equal(t, "req-1", entry["request_id"])
	assert.Equal(t, "203.0.113.7", entry["client_ip"])
	assert.Equal(t, "test-agent", entry["user_agent"])
	assert.NotContains(t, entry, "reason", "no reason is logged when none was given")
}

// On an authenticated request the middleware has already established who is calling, so an
// event does not have to repeat it.
func TestSecurity_FallsBackToContextIdentity(t *testing.T) {
	entries := captureSecurity(t, func() {
		ctx := context.WithValue(requestContext(), constants.ContextKeyUserID, "ctx-user")
		ctx = context.WithValue(ctx, constants.ContextKeySessionID, "ctx-session")

		Security(ctx, SecurityEvent{Action: ActionSessionRevoked, Outcome: OutcomeSuccess})
	})

	require.Len(t, entries, 1)
	assert.Equal(t, "ctx-user", entries[0]["user_id"])
	assert.Equal(t, "ctx-session", entries[0]["session_id"])
}

// Login is the case that matters: the context has no identity yet, so an explicitly passed one
// must win rather than be dropped.
func TestSecurity_ExplicitIdentityOverridesContext(t *testing.T) {
	entries := captureSecurity(t, func() {
		ctx := context.WithValue(requestContext(), constants.ContextKeyUserID, "ctx-user")

		Security(ctx, SecurityEvent{
			Action:  ActionLoginSucceeded,
			Outcome: OutcomeSuccess,
			UserID:  "explicit-user",
		})
	})

	require.Len(t, entries, 1)
	assert.Equal(t, "explicit-user", entries[0]["user_id"])
}

// A failure has to clear the default INFO threshold, or the events an operator most wants —
// the rejected ones — would be the only ones not written.
func TestSecurity_FailuresAreWarnings(t *testing.T) {
	entries := captureSecurity(t, func() {
		Security(requestContext(), SecurityEvent{
			Action:  ActionLoginFailed,
			Outcome: OutcomeFailure,
			Reason:  ReasonBadPassword,
		})
		Security(requestContext(), SecurityEvent{
			Action:  ActionLoginSucceeded,
			Outcome: OutcomeSuccess,
		})
	})

	require.Len(t, entries, 2)
	assert.Equal(t, slog.LevelWarn.String(), entries[0]["level"])
	assert.Equal(t, ReasonBadPassword, entries[0]["reason"])
	assert.Equal(t, slog.LevelInfo.String(), entries[1]["level"])
}

// Nothing may break when there is no HTTP request behind the call — a cleanup job or a gRPC
// handler has no client IP, and that must not stop the event being recorded.
func TestSecurity_ToleratesMissingContextValues(t *testing.T) {
	entries := captureSecurity(t, func() {
		Security(context.Background(), SecurityEvent{Action: ActionAccountDeactivated, Outcome: OutcomeSuccess, UserID: "u1"})
	})

	require.Len(t, entries, 1)
	assert.Equal(t, "", entries[0]["client_ip"])
	assert.Equal(t, "", entries[0]["request_id"])
	assert.Equal(t, "u1", entries[0]["user_id"])
}
