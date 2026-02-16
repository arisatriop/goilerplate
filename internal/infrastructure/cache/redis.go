package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService provides generic Redis operations
type RedisService struct {
	client  *redis.Client
	enabled bool
}

// NewRedisService creates a new Redis service
func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client:  client,
		enabled: client != nil,
	}
}

// IsEnabled returns whether Redis is enabled
func (s *RedisService) IsEnabled() bool {
	return s.enabled
}

// Get retrieves a raw string value from Redis
// Returns (value, error) - returns redis.Nil if key not found
func (s *RedisService) Get(ctx context.Context, key string) (string, error) {
	if !s.enabled {
		return "", nil
	}

	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return result, nil
}

// Set stores a raw string value in Redis with TTL
func (s *RedisService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if !s.enabled {
		return nil // Skip when Redis is disabled
	}

	if err := s.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}

	return nil
}

// GetJSON retrieves a value from Redis and unmarshals it into result
// Returns nil if found, redis.Nil if not found, error for other issues
func (s *RedisService) GetJSON(ctx context.Context, key string, result interface{}) error {
	if !s.enabled {
		return redis.Nil // Treat as cache miss when disabled
	}

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// SetJSON marshals value to JSON and stores it in Redis with TTL
func (s *RedisService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !s.enabled {
		return nil // Skip when Redis is disabled
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes one or more keys from Redis
func (s *RedisService) Delete(ctx context.Context, keys ...string) error {
	if !s.enabled {
		return nil // Skip when Redis is disabled
	}

	if len(keys) == 0 {
		return nil
	}

	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (s *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	if !s.enabled {
		return false, nil
	}

	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// DeleteByPattern deletes all keys matching a pattern
func (s *RedisService) DeleteByPattern(ctx context.Context, pattern string) error {
	if !s.enabled {
		return nil // Skip when Redis is disabled
	}

	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		s.client.Del(ctx, key)
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan pattern %s: %w", pattern, err)
	}

	return nil
}

// GetClient returns the underlying Redis client for advanced operations
func (s *RedisService) GetClient() *redis.Client {
	return s.client
}
