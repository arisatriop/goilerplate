package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func CreateRedisConnection() *redis.Client {

	app := GetAppVariable()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     app.RedisHost + ":" + app.RedisPort,
		Password: app.RedisPassword,
		DB:       0,
	})

	ctx := context.Background()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to redis: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Redis: connected")
	fmt.Println()

	return GetRedisConnection()
}

func GetRedisConnection() *redis.Client {
	return redisClient
}
