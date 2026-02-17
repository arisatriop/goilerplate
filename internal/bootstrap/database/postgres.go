package bootstrap

import (
	"context"
	"fmt"
	"goilerplate/config"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg *config.Config, log *slog.Logger) *pgxpool.Pool {

	var pgx *pgxpool.Pool

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {

		log.Error("Unable to parse postgres config", "error", err)
		panic(fmt.Sprintf("failed to parse postgres config: %v", err))
	}

	config.MinConns = int32(cfg.DB.MinOpenConnections)
	config.MaxConns = int32(cfg.DB.MaxOpenConnections)
	config.MaxConnLifetime = time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime)
	config.MaxConnIdleTime = time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime)
	config.HealthCheckPeriod = time.Second * time.Duration(cfg.DB.HealthCheckPeriod)

	if pgx, err = pgxpool.NewWithConfig(context.Background(), config); err != nil {
		log.Error(fmt.Sprintf("failed to create postgres pool: %v", err))
		os.Exit(1)
	}

	if err = pgx.Ping(context.Background()); err != nil {
		log.Error(fmt.Sprintf("failed to ping postgres: %v", err))
		os.Exit(1)
	}

	return pgx
}
