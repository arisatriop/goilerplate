package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func CreateRedisConnection() {

	app := GetAppVariable()

	rdb := redis.NewClient(&redis.Options{
		Addr:     app.RedisHost.(string) + ":" + app.RedisPort.(string),
		Password: app.RedisPassword.(string),
		DB:       0,
	})

	ctx := context.Background()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to redis: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Redis: connected")
	fmt.Println()
}
