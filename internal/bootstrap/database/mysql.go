package bootstrap

import (
	"database/sql"
	"fmt"
	"goilerplate/config"
	"log/slog"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func NewMysql(cfg *config.Config, log *slog.Logger) *sql.DB {

	var db *sql.DB
	var err error

	config := mysql.Config{
		User:                 cfg.DB.Username,
		Passwd:               cfg.DB.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port),
		DBName:               cfg.DB.Name,
		AllowNativePasswords: true,
	}

	if db, err = sql.Open("mysql", config.FormatDSN()); err != nil {
		log.Error(fmt.Sprintf("failed to connect to mysql: %v", err))
		os.Exit(1)
	}
	if err = db.Ping(); err != nil {
		log.Error(fmt.Sprintf("failed to ping mysql: %v", err))
		os.Exit(1)
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	db.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	db.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return db
}
