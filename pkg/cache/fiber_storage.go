package cache

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// FiberStorage implements fiber.Storage over the existing go-redis client.
// All keys are namespaced under the given prefix so Reset() only deletes
// keys belonging to this storage instance, not the whole Redis DB.
type FiberStorage struct {
	client *redis.Client
	prefix string
}

// NewFiberStorage returns a fiber.Storage backed by Redis.
// Returns nil if client is nil, so the caller can pass it directly to
// fiber's limiter.Config.Storage (nil = in-memory fallback).
func NewFiberStorage(client *redis.Client, prefix string) fiber.Storage {
	if client == nil {
		return nil
	}
	return &FiberStorage{client: client, prefix: prefix}
}

func (s *FiberStorage) Get(key string) ([]byte, error) {
	val, err := s.client.Get(context.Background(), s.prefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return val, err
}

func (s *FiberStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.client.Set(context.Background(), s.prefix+key, val, exp).Err()
}

func (s *FiberStorage) Delete(key string) error {
	return s.client.Del(context.Background(), s.prefix+key).Err()
}

// Reset deletes only keys belonging to this prefix, not the full Redis DB.
func (s *FiberStorage) Reset() error {
	ctx := context.Background()
	iter := s.client.Scan(ctx, 0, s.prefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

func (s *FiberStorage) Close() error {
	return nil
}
