package db

import (
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/sirupsen/logrus"
	gormMysql "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func gormDB(driver string, cfg *config.Config, log *logrus.Logger) *gorm.DB {
	var dialector gorm.Dialector
	usename := cfg.DB.Username
	password := cfg.DB.Password
	host := cfg.DB.Host
	port := cfg.DB.Port
	dbName := cfg.DB.Name

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
		log.Fatalf("unsupported db driver: %s", driver)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Warn,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect to gorm: %v", err)
	}

	connection, err := gdb.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	connection.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	connection.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return gdb
}
