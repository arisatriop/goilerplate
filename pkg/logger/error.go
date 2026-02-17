package logger

import (
	"context"
	"goilerplate/pkg/constants"
	"log/slog"
)

const (
	LogLabel = "application-log"
)

// Error logs an error message with context information such as request ID and user details.
func Error(ctx context.Context, err error) {
	var requestID, userID, userName string

	if val := ctx.Value(constants.ContextKeyRequestID); val != nil {
		if id, ok := val.(string); ok {
			requestID = id
		}
	}

	if val := ctx.Value(constants.ContextKeyUserID); val != nil {
		if id, ok := val.(string); ok {
			userID = id
		}
	}

	if val := ctx.Value(constants.ContextKeyUserName); val != nil {
		if name, ok := val.(string); ok {
			userName = name
		}
	}

	logAttrs := []slog.Attr{
		slog.String("label", LogLabel),
		slog.String("request_id", requestID),
		slog.String("user_id", userID),
		slog.String("user_name", userName),
		slog.Any("error", err),
	}

	// Always log as ERROR
	slog.LogAttrs(ctx, slog.LevelError, "Application Log", logAttrs...)
}
