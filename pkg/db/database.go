package db

import (
	"database/sql"
	"goilerplate/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DB struct {
	SqlDB   *sql.DB
	PgxPool *pgxpool.Pool
	GDB     *gorm.DB
	Redis   *redis.Client
}

func NewDatabase(cfg *config.Config, log *logrus.Logger) *DB {
	return &DB{
		// SqlDB:   mysqlDB(cfg, log),
		PgxPool: postgresDB(cfg, log),
		GDB:     gormDB("postgres", cfg, log),
		Redis:   redisDB(cfg, log),
	}
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
