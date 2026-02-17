package context

import (
	"context"

	"goilerplate/pkg/constants"
)

// AuditInfo contains information for audit fields
type AuditInfo struct {
	UserID   string
	UserName string
}

// WithAuditInfo adds audit information to context
// NOTE: For authenticated requests, auth middleware already sets this.
// Only use this for unauthenticated operations (e.g., registration, system tasks)
func WithAuditInfo(ctx context.Context, userID, userName string) context.Context {
	ctx = context.WithValue(ctx, constants.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, constants.ContextKeyUserName, userName)
	return ctx
}

// GetUserID extracts user ID from context for audit purposes
// Works with both auth middleware context and manually set context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(constants.ContextKeyUserID).(string); ok {
		return userID
	}
	return "system" // default fallback
}

// GetUserName extracts user name from context for audit purposes
// Works with both auth middleware context and manually set context
func GetUserName(ctx context.Context) string {
	if userName, ok := ctx.Value(constants.ContextKeyUserName).(string); ok {
		return userName
	}
	return "system" // default fallback
}

// GetAuditInfo extracts complete audit information from context
func GetAuditInfo(ctx context.Context) AuditInfo {
	return AuditInfo{
		UserID:   GetUserID(ctx),
		UserName: GetUserName(ctx),
	}
}
