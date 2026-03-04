package logger

import (
	"context"
	"errors"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
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

// logWithSource logs with an explicit source location
func logWithSource(ctx context.Context, level slog.Level, msg, source string) {
	info := extractContext(ctx)
	attrs := append(info.baseAttrs(),
		slog.String("source", source),
		slog.Any("message", msg),
	)
	slog.LogAttrs(ctx, level, "Application Log", attrs...)
}

// Error logs an error message with context information such as request ID and user details.
// If the error is an InternalError, it extracts and logs the original file:line location.
func Error(ctx context.Context, err error) {
	var internalErr *utils.InternalError
	if errors.As(err, &internalErr) {
		logWithSource(ctx, slog.LevelError, err.Error(), internalErr.Location())
		return
	}
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
