package bootstrap

import (
	"context"
	"fmt"
	"goilerplate/config"

	"github.com/redis/go-redis/v9"
)

// NewRedis connects to Redis when redis.enabled is set, and returns nil otherwise. The ping is
// bounded by the dial timeout, so an unreachable server fails startup instead of hanging it.
func NewRedis(cfg *config.Config) (*redis.Client, error) {
	if !cfg.Redis.Enabled {
		return nil, nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  cfg.Redis.PoolTimeout,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close() // the client is discarded; its close error adds nothing to the ping's
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	return rdb, nil
}
