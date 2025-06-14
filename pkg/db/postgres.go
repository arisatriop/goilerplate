package db

import (
	"context"
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func postgresDB(cfg *config.Config, log *logrus.Logger) *pgxpool.Pool {

	var pgx *pgxpool.Pool

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", "err")
	}

	config.MinConns = int32(cfg.DB.MinOpenConnections)
	config.MaxConns = int32(cfg.DB.MaxOpenConnections)
	config.MaxConnLifetime = time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime)
	config.MaxConnIdleTime = time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime)
	config.HealthCheckPeriod = time.Second * time.Duration(cfg.DB.HealthCheckPeriod)

	if pgx, err = pgxpool.NewWithConfig(context.Background(), config); err != nil {
		log.Fatalf("Unable to connect to postgres: %v", err)
	}

	if err = pgx.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	return pgx
}
