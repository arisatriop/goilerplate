package pkg

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	Ctx    context.Context
	Client *redis.Client
}

func NewRedis(r *redis.Client, ctx context.Context) *redisClient {
	return &redisClient{
		Ctx:    ctx,
		Client: r,
	}
}

// Set sets the value of a key with an optional TTL.
func (c *redisClient) Set(key string, value any, ttl time.Duration) error {
	return c.Client.Set(c.Ctx, key, value, ttl).Err()
}

// Get returns the string value of a key.
func (c *redisClient) Get(key string) (string, error) {
	return c.Client.Get(c.Ctx, key).Result()
}

// Del deletes a key.
func (c *redisClient) Del(key string) error {
	return c.Client.Del(c.Ctx, key).Err()
}

// Exists checks if the key exists.
func (c *redisClient) Exists(key string) (bool, error) {
	res, err := c.Client.Exists(c.Ctx, key).Result()
	return res > 0, err
}

// SetNX sets a key if it does not exist.
func (c *redisClient) SetNX(key string, value interface{}, ttl time.Duration) (bool, error) {
	return c.Client.SetNX(c.Ctx, key, value, ttl).Result()
}

// TTL returns the time-to-live of a key.
func (c *redisClient) TTL(key string) (time.Duration, error) {
	return c.Client.TTL(c.Ctx, key).Result()
}
