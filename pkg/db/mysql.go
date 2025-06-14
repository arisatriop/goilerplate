package db

import (
	"database/sql"
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func mysqlDB(cfg *config.Config, log *logrus.Logger) *sql.DB {

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
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping mysql: %v", err)
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	db.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	db.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return db
}
