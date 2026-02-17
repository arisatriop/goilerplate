package bootstrap

import (
	"fmt"
	"goilerplate/config"
	"log/slog"
	"os"
	"strings"
	"time"

	gormMysql "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGorm(cfg *config.Config, log *slog.Logger) *gorm.DB {
	var dialector gorm.Dialector
	usename := cfg.DB.Username
	password := cfg.DB.Password
	host := cfg.DB.Host
	port := cfg.DB.Port
	dbName := cfg.DB.Name

	driver := strings.ToLower(cfg.DB.Driver)

	switch driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			host,
			usename,
			password,
			dbName,
			port,
		)
		dialector = gormPostgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&allowNativePasswords=true",
			usename,
			password,
			host,
			port,
			dbName,
		)
		dialector = gormMysql.Open(dsn)
	default:
		slog.Error(fmt.Sprintf("failed to connect to gorm: unsupported db driver %s", driver))
		os.Exit(1)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
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
