package db

import (
	"context"
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func redisDB(cfg *config.Config, log *logrus.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password, // no password set
		DB:           cfg.Redis.DB,       // use default DB
		DialTimeout:  time.Second * time.Duration(cfg.Redis.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(cfg.Redis.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(cfg.Redis.WriteTimeout),
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  time.Second * time.Duration(cfg.Redis.PoolTimeout),
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	return rdb
}
