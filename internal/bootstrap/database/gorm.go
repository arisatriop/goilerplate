package bootstrap

import (
	"fmt"
	"goilerplate/config"
	"goilerplate/pkg/utils"
	"log/slog"
	"os"
	"time"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

func NewGorm(cfg *config.Config, log *slog.Logger) *gorm.DB {
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
		log.Error(fmt.Sprintf("failed to connect to gorm: %v", err))
		os.Exit(1)
	}

	if cfg.OTel.Enabled {
		if err := gdb.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			log.Error(fmt.Sprintf("failed to register GORM OTel plugin: %v", err))
		}
	}

	connection, err := gdb.DB()
	if err != nil {
		log.Error(fmt.Sprintf("failed to get sql.DB from gorm: %v", err))
		os.Exit(1)
	}

	connection.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	connection.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return gdb
}
