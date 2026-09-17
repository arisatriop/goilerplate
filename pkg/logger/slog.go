package logger

import (
	"goilerplate/config"
	"goilerplate/pkg/redact"
	"log/slog"
	"os"
	"strings"
)

func NewSlog(cfg *config.Config) *slog.Logger {
	logLevel := slog.LevelInfo
	logSource := false

	if cfg != nil && cfg.Log != nil {
		logLevel = getLogLevel(cfg.Log.Level)
	}
	if cfg != nil && cfg.Log != nil {
		logSource = cfg.Log.Source
		redact.SetDefault(redact.New(cfg.Log.RedactFields...))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logSource,
	}))

	slog.SetDefault(logger)

	return logger
}

func getLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
