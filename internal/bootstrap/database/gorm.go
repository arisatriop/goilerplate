package bootstrap

import (
	"fmt"
	"goilerplate/config"
	"goilerplate/pkg/utils"
	"log/slog"
	"time"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

// NewGorm opens the PostgreSQL pool. gorm.Open pings the server, so an unreachable database
// fails here, at startup, rather than on the first request.
func NewGorm(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	dialector := gormPostgres.Open(PostgresDSN(cfg.DB))

	gdb, err := gorm.Open(dialector, &gorm.Config{
		NowFunc:                utils.Now,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		QueryFields:            true,
		Logger: logger.New(NewSlogWriter(log), logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Warn,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if cfg.OTel.Enabled {
		if err := gdb.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			// Tracing is optional: without the plugin queries still run, they are just not traced.
			log.Error("registering GORM OTel plugin", "error", err)
		}
	}

	connection, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("getting sql.DB from gorm: %w", err)
	}

	connection.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	// db.min_open_connections is the pool's warm floor. database/sql spells it as the idle
	// count: it keeps that many connections open rather than closing them after each use, so
	// a burst of traffic does not pay for a TCP handshake and a TLS negotiation per request.
	connection.SetMaxIdleConns(cfg.DB.MinOpenConnections)
	connection.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return gdb, nil
}
