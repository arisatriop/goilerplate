package bootstrap

import (
	"fmt"
	"log/slog"
	"strings"

	"goilerplate/config"

	"gorm.io/gorm"
)

// DB holds the PostgreSQL connection.
//
// There is one pool. The process used to open a second, standalone pgxpool alongside GORM; it
// was pinged by the health check and never queried, so a deployment sized for 20 connections
// quietly consumed 40. GORM reaches PostgreSQL through the pgx stdlib driver either way.
type DB struct {
	GDB *gorm.DB
}

func NewDB() *DB {
	return &DB{}
}

// PostgresDSN builds a libpq key/value connection string shared by pgx and GORM.
// Values are quoted, so passwords may contain spaces, quotes, or backslashes.
// The session time zone is UTC, so timestamps never depend on the database server setting.
func PostgresDSN(db config.DB) string {
	return strings.Join([]string{
		"host=" + quoteDSNValue(db.Host),
		fmt.Sprintf("port=%d", db.Port),
		"user=" + quoteDSNValue(db.Username),
		"password=" + quoteDSNValue(db.Password),
		"dbname=" + quoteDSNValue(db.Name),
		"sslmode=" + quoteDSNValue(sslModeOrDefault(db.SSLMode)),
		// Unquoted: gorm.io/driver/postgres reads timezone= from the DSN and passes the raw value on
		"timezone=UTC",
	}, " ")
}

func quoteDSNValue(value string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value) + "'"
}

// sslModeOrDefault keeps libpq's own default ("prefer") when db.sslmode is unset.
func sslModeOrDefault(mode string) string {
	if mode == "" {
		return "prefer"
	}
	return mode
}

type slogWriter struct {
	Logger *slog.Logger
	Level  slog.Level
}

func (s *slogWriter) Printf(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)

	switch s.Level {
	case slog.LevelDebug:
		s.Logger.Debug(formattedMessage)
	case slog.LevelInfo:
		s.Logger.Info(formattedMessage)
	case slog.LevelWarn:
		s.Logger.Warn(formattedMessage)
	case slog.LevelError:
		s.Logger.Error(formattedMessage)
	default:
		s.Logger.Debug(formattedMessage)
	}
}

func NewSlogWriter(logger *slog.Logger) *slogWriter {
	return &slogWriter{
		Logger: logger,
		Level:  slog.LevelDebug,
	}
}
