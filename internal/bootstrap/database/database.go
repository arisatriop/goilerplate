package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

const (
	Postgres = "postgres"
	Mysql    = "mysql"
)

type DB struct {
	PgxDB   *pgxpool.Pool
	MysqlDB *sql.DB
	GDB     *gorm.DB
}

func NewDB() *DB {
	return &DB{}
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

func NewSlogWriterWithLevel(logger *slog.Logger, level slog.Level) *slogWriter {
	return &slogWriter{
		Logger: logger,
		Level:  level,
	}
}
