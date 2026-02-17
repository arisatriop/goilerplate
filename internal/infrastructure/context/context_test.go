package context_test

import (
	"context"
	"testing"

	auditctx "goilerplate/internal/infrastructure/context"
	"goilerplate/pkg/constants"
)

func TestAuditContext_WithAndWithoutTransaction(t *testing.T) {
	t.Run("should set and get user ID from context", func(t *testing.T) {
		ctx := context.Background()
		ctx = auditctx.WithAuditInfo(ctx, "user123", "John Doe")

		userID := auditctx.GetUserID(ctx)
		userName := auditctx.GetUserName(ctx)

		if userID != "user123" {
			t.Errorf("Expected userID 'user123', got '%s'", userID)
		}
		if userName != "John Doe" {
			t.Errorf("Expected userName 'John Doe', got '%s'", userName)
		}
	})

	t.Run("should use same context keys as auth middleware", func(t *testing.T) {
		ctx := context.Background()

		// Simulate auth middleware setting context
		ctx = context.WithValue(ctx, constants.ContextKeyUserID, "middleware-user")
		ctx = context.WithValue(ctx, constants.ContextKeyUserName, "Middleware User")

		// Audit context should read the same keys
		userID := auditctx.GetUserID(ctx)
		userName := auditctx.GetUserName(ctx)

		if userID != "middleware-user" {
			t.Errorf("Expected userID 'middleware-user', got '%s'", userID)
		}
		if userName != "Middleware User" {
			t.Errorf("Expected userName 'Middleware User', got '%s'", userName)
		}
	})

	t.Run("should return 'system' as fallback", func(t *testing.T) {
		ctx := context.Background()

		userID := auditctx.GetUserID(ctx)
		userName := auditctx.GetUserName(ctx)

		if userID != "system" {
			t.Errorf("Expected fallback userID 'system', got '%s'", userID)
		}
		if userName != "system" {
			t.Errorf("Expected fallback userName 'system', got '%s'", userName)
		}
	})
}
