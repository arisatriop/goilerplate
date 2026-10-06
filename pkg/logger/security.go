package logger

import (
	"context"
	"goilerplate/pkg/constants"
	"log/slog"
)

// SecurityLabel marks the security audit trail. It is deliberately different from LogLabel so a
// SIEM can select these events with one filter, without having to recognise message text that
// may be reworded at any time.
const SecurityLabel = "security-log"

// Actions name what happened. They are a closed vocabulary: alerting rules are written against
// these strings, so a rule stops matching the moment one is renamed. Treat them as an API.
const (
	ActionLoginSucceeded     = "login_succeeded"
	ActionLoginFailed        = "login_failed"
	ActionAccountLocked      = "account_locked"
	ActionTokenReuseDetected = "refresh_token_reuse_detected"
	ActionSessionRevoked     = "session_revoked"
	ActionAllSessionsRevoked = "all_sessions_revoked"
	ActionPasswordChanged    = "password_changed"
	ActionAccountDeactivated = "account_deactivated"

	ActionPasswordResetRequested = "password_reset_requested"
	ActionPasswordResetCompleted = "password_reset_completed"

	// Emitted once the email change flow exists. Named here so it adopts the vocabulary
	// rather than inventing a second one.
	ActionEmailChanged = "email_changed"
)

// Outcomes. A failure is logged at WARN so that the default INFO level still shows every
// rejected attempt, which is what an operator actually wants to be woken up by.
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)

// Reasons explain a failure. They describe the server's decision, never the credential that
// caused it: a reason string must stay safe to ship to a log aggregator.
const (
	ReasonUnknownEmail    = "unknown_email"
	ReasonBadPassword     = "bad_password"
	ReasonAccountLocked   = "account_locked"
	ReasonAccountDisabled = "account_disabled"
	// ReasonCooldown: a password reset was asked for again too soon; nothing was sent.
	ReasonCooldown = "cooldown"
	// ReasonDeliveryFailed: the reset email could not be queued.
	ReasonDeliveryFailed = "delivery_failed"
	// ReasonInvalidToken: a one-time token was unknown, used, superseded or expired.
	ReasonInvalidToken = "invalid_token"
)

// SecurityEvent is one entry in the audit trail.
//
// UserID and SessionID may be left empty, in which case they are read from the context. That
// covers authenticated requests, where the auth middleware has already put them there. Login is
// the case that must set them explicitly: it establishes the identity the context does not have
// yet, and a failed login has no session at all.
type SecurityEvent struct {
	Action    string
	Outcome   string
	UserID    string
	SessionID string
	Reason    string
}

// Security writes one structured audit entry.
//
// Every field is either an identifier or a fixed vocabulary word. Nothing derived from a
// credential — no password, no token, not even a token ID — belongs here: the audit trail is
// routinely shipped off-host and kept far longer than any secret's lifetime.
func Security(ctx context.Context, event SecurityEvent) {
	level := slog.LevelInfo
	if event.Outcome == OutcomeFailure {
		level = slog.LevelWarn
	}

	attrs := []slog.Attr{
		slog.String("label", SecurityLabel),
		slog.String("request_id", stringFromContext(ctx, constants.ContextKeyRequestID)),
		slog.String("action", event.Action),
		slog.String("outcome", event.Outcome),
		slog.String("user_id", orContext(ctx, event.UserID, constants.ContextKeyUserID)),
		slog.String("session_id", orContext(ctx, event.SessionID, constants.ContextKeySessionID)),
		slog.String("client_ip", stringFromContext(ctx, constants.ContextKeyClientIP)),
		slog.String("user_agent", stringFromContext(ctx, constants.ContextKeyUserAgent)),
	}
	if event.Reason != "" {
		attrs = append(attrs, slog.String("reason", event.Reason))
	}

	slog.LogAttrs(ctx, level, "Security event", attrs...)
}

// orContext prefers an explicitly supplied value and falls back to the context.
func orContext(ctx context.Context, value string, key constants.ContextKey) string {
	if value != "" {
		return value
	}
	return stringFromContext(ctx, key)
}

// stringFromContext reads a string context value, returning "" when it is absent or is not a
// string. A missing value is normal — background jobs and gRPC calls have no HTTP context — so
// it is not an error.
func stringFromContext(ctx context.Context, key constants.ContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}
