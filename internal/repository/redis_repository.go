package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		Client: client,
	}
}

// Set sets a key with a given value and optional TTL.
func (r *RedisRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Get returns the string value of a key.
func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del deletes a key.
func (r *RedisRepository) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// Exists checks if the key exists.
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	res, err := r.Client.Exists(ctx, key).Result()
	return res > 0, err
}

// SetNX sets a key if it does not exist.
func (r *RedisRepository) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, ttl).Result()
}

// TTL returns the time-to-live of a key.
func (r *RedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Close closes the Redis connection.
func (r *RedisRepository) Close() error {
	return r.Client.Close()
}
