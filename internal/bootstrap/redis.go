package bootstrap

import (
	"context"
	"fmt"
	"goilerplate/config"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config, log *slog.Logger) *redis.Client {
	if cfg.Redis.Enabled {

		rdb := redis.NewClient(&redis.Options{
			Addr:         cfg.Redis.Host,
			Password:     cfg.Redis.Password, // no password set
			DB:           cfg.Redis.DB,       // use default DB
			DialTimeout:  cfg.Redis.DialTimeout,
			ReadTimeout:  cfg.Redis.ReadTimeout,
			WriteTimeout: cfg.Redis.WriteTimeout,
			PoolSize:     cfg.Redis.PoolSize,
			PoolTimeout:  cfg.Redis.PoolTimeout,
		})

		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Error(fmt.Sprintf("failed to connect to redis: %v", err))
			os.Exit(1)
		}

		return rdb
	}

	return nil
}
