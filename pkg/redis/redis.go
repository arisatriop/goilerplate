package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func Set(r *redis.Client, ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.Set(ctx, key, value, ttl).Err()
}

// Get returns the string value of a key.
func Get(r *redis.Client, ctx context.Context, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

// Del deletes a key.
func Del(r *redis.Client, ctx context.Context, key string) error {
	return r.Del(ctx, key).Err()
}

// Exists checks if the key exists.
func Exists(r *redis.Client, ctx context.Context, key string) (bool, error) {
	res, err := r.Exists(ctx, key).Result()
	return res > 0, err
}

// SetNX sets a key if it does not exist.
func SetNX(r *redis.Client, ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return r.SetNX(ctx, key, value, ttl).Result()
}

// TTL returns the time-to-live of a key.
func TTL(r *redis.Client, ctx context.Context, key string) (time.Duration, error) {
	return r.TTL(ctx, key).Result()
}
