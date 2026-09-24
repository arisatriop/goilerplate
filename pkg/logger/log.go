package logger

import (
	"context"
	"goilerplate/pkg/constants"
	"log/slog"
)

const (
	LogLabel = "application-log"
)

// contextInfo holds extracted context values for logging
type contextInfo struct {
	requestID string
	userID    string
	userName  string
}

// extractContext extracts common context values used in logging
func extractContext(ctx context.Context) contextInfo {
	info := contextInfo{}

	if val := ctx.Value(constants.ContextKeyRequestID); val != nil {
		if id, ok := val.(string); ok {
			info.requestID = id
		}
	}

	if val := ctx.Value(constants.ContextKeyUserID); val != nil {
		if id, ok := val.(string); ok {
			info.userID = id
		}
	}

	if val := ctx.Value(constants.ContextKeyUserName); val != nil {
		if name, ok := val.(string); ok {
			info.userName = name
		}
	}

	return info
}

// baseAttrs returns the common log attributes
func (c contextInfo) baseAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("label", LogLabel),
		slog.String("request_id", c.requestID),
		slog.String("user_id", c.userID),
		slog.String("user_name", c.userName),
	}
}

func Log(ctx context.Context, level slog.Level, msg string) {
	info := extractContext(ctx)
	attrs := append(info.baseAttrs(), slog.Any("message", msg))
	slog.LogAttrs(ctx, level, "Application Log", attrs...)
}

// Error logs err at error level with the request's context. The message is the whole wrap chain,
// so the context each layer added ("inserting bar: ...") says where the failure came from.
func Error(ctx context.Context, err error) {
	Log(ctx, slog.LevelError, err.Error())
}

func Warn(ctx context.Context, msg string) {
	Log(ctx, slog.LevelWarn, msg)
}

func Info(ctx context.Context, msg string) {
	Log(ctx, slog.LevelInfo, msg)
}

func Debug(ctx context.Context, msg string) {
	Log(ctx, slog.LevelDebug, msg)
}
