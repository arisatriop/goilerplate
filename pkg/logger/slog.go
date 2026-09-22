package logger

import (
	"log/slog"
	"os"
	"strings"

	"goilerplate/pkg/redact"
)

// Options configures the process logger.
//
// It is a plain struct rather than *config.Config on purpose. pkg/logger is a general-purpose
// package; importing this application's configuration struct made it unusable anywhere else and
// pointed the dependency the wrong way — a shared utility should not know what the application
// it serves keeps in its YAML. The caller in internal/bootstrap does the translation.
type Options struct {
	// Level is DEBUG, INFO, WARN or ERROR, case-insensitive. Anything else means INFO.
	Level string

	// AddSource attaches the file and line that emitted each record. Useful in development,
	// measurable overhead per record in production.
	AddSource bool

	// RedactFields are attribute keys whose values are replaced before they are written.
	RedactFields []string
}

// New builds the JSON logger the process writes to stdout and installs it as slog's default,
// so a package that logs without being handed a logger still writes in the same format.
func New(opts Options) *slog.Logger {
	redact.SetDefault(redact.New(opts.RedactFields...))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     parseLevel(opts.Level),
		AddSource: opts.AddSource,
	}))

	slog.SetDefault(logger)

	return logger
}

func parseLevel(level string) slog.Level {
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
