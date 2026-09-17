package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/lock"

	"github.com/redis/go-redis/v9"
)

const (
	defaultAuthKeyPrefix = "auth:"
	defaultLockKeyPrefix = "lock:"
	redisDeleteBatchSize = 500
)

// RedisSessionStore is an auth.SessionStore shared by every instance.
// Sessions are indexed per user in a set, so DeleteByUser never scans the keyspace.
type RedisSessionStore struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
	now    func() time.Time
}

// NewRedisSessionStore creates a RedisSessionStore that keeps sessions for at most ttl.
func NewRedisSessionStore(client *redis.Client, ttl time.Duration) *RedisSessionStore {
	return &RedisSessionStore{client: client, ttl: ttl, prefix: defaultAuthKeyPrefix, now: time.Now}
}

func (s *RedisSessionStore) sessionKey(sessionID string) string {
	return s.prefix + "session:" + sessionID
}

func (s *RedisSessionStore) userKey(userID string) string {
	return s.prefix + "user_sessions:" + userID
}

func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (*auth.UserSession, bool, error) {
	data, err := s.client.Get(ctx, s.sessionKey(sessionID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("getting cached session: %w", err)
	}

	var session auth.UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, false, fmt.Errorf("decoding cached session: %w", err)
	}
	return &session, true, nil
}

func (s *RedisSessionStore) Set(ctx context.Context, session *auth.UserSession) error {
	ttl := sessionTTL(session, s.ttl, s.now())
	if ttl <= 0 {
		return nil
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("encoding session: %w", err)
	}

	// The index outlives its members because every member TTL is at most s.ttl.
	userKey := s.userKey(session.UserID)
	_, err = s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, s.sessionKey(session.ID), data, ttl)
		pipe.SAdd(ctx, userKey, session.ID)
		pipe.Expire(ctx, userKey, s.ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("caching session: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	if err := s.client.Del(ctx, s.sessionKey(sessionID)).Err(); err != nil {
		return fmt.Errorf("deleting cached session: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) DeleteByUser(ctx context.Context, userID string) error {
	userKey := s.userKey(userID)
	if err := deleteSetMembers(ctx, s.client, userKey, s.sessionKey); err != nil {
		return fmt.Errorf("deleting cached sessions of user: %w", err)
	}
	return nil
}

// RedisPermissionCache is an auth.PermissionCache shared by every instance.
// Cached users are tracked in a set, so InvalidateAll never scans the keyspace.
type RedisPermissionCache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

// NewRedisPermissionCache creates a RedisPermissionCache that keeps permissions for ttl.
func NewRedisPermissionCache(client *redis.Client, ttl time.Duration) *RedisPermissionCache {
	return &RedisPermissionCache{client: client, ttl: ttl, prefix: defaultAuthKeyPrefix}
}

func (c *RedisPermissionCache) permissionKey(userID string) string {
	return c.prefix + "permissions:" + userID
}

func (c *RedisPermissionCache) indexKey() string {
	return c.prefix + "permission_users"
}

func (c *RedisPermissionCache) Get(ctx context.Context, userID string) ([]string, bool, error) {
	data, err := c.client.Get(ctx, c.permissionKey(userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("getting cached permissions: %w", err)
	}

	var permissions []string
	if err := json.Unmarshal(data, &permissions); err != nil {
		return nil, false, fmt.Errorf("decoding cached permissions: %w", err)
	}
	return permissions, true, nil
}

func (c *RedisPermissionCache) Set(ctx context.Context, userID string, permissions []string) error {
	if c.ttl <= 0 {
		return nil
	}
	if permissions == nil {
		permissions = []string{}
	}

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("encoding permissions: %w", err)
	}

	_, err = c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, c.permissionKey(userID), data, c.ttl)
		pipe.SAdd(ctx, c.indexKey(), userID)
		pipe.Expire(ctx, c.indexKey(), c.ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("caching permissions: %w", err)
	}
	return nil
}

func (c *RedisPermissionCache) Invalidate(ctx context.Context, userID string) error {
	if err := c.client.Del(ctx, c.permissionKey(userID)).Err(); err != nil {
		return fmt.Errorf("invalidating cached permissions: %w", err)
	}
	return nil
}

func (c *RedisPermissionCache) InvalidateAll(ctx context.Context) error {
	if err := deleteSetMembers(ctx, c.client, c.indexKey(), c.permissionKey); err != nil {
		return fmt.Errorf("invalidating all cached permissions: %w", err)
	}
	return nil
}

// deleteSetMembers deletes the key of every member of setKey in batches, then the set.
// An entry added concurrently may survive; it still expires with its own TTL.
func deleteSetMembers(ctx context.Context, client *redis.Client, setKey string, memberKey func(string) string) error {
	iter := client.SScan(ctx, setKey, 0, "", redisDeleteBatchSize).Iterator()
	batch := make([]string, 0, redisDeleteBatchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		_, err := client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, key := range batch {
				pipe.Del(ctx, key)
			}
			return nil
		})
		batch = batch[:0]
		return err
	}

	for iter.Next(ctx) {
		batch = append(batch, memberKey(iter.Val()))
		if len(batch) == redisDeleteBatchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if err := flush(); err != nil {
		return err
	}
	return client.Del(ctx, setKey).Err()
}

// RedisLocker is a lock.Provider shared by every instance (SET NX with an owner token).
type RedisLocker struct {
	client *redis.Client
	prefix string
}

// NewRedisLocker creates a RedisLocker.
func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{client: client, prefix: defaultLockKeyPrefix}
}

// releaseScript deletes the lock only if it is still held by the caller's token.
var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

func (l *RedisLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (lock.Release, bool, error) {
	token, err := randomToken()
	if err != nil {
		return nil, false, fmt.Errorf("generating lock token: %w", err)
	}

	lockKey := l.prefix + key
	acquired, err := l.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("acquiring lock: %w", err)
	}
	if !acquired {
		return nil, false, nil
	}

	release := func(ctx context.Context) error {
		if err := releaseScript.Run(ctx, l.client, []string{lockKey}, token).Err(); err != nil {
			return fmt.Errorf("releasing lock: %w", err)
		}
		return nil
	}
	return release, true, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var (
	_ auth.SessionStore    = (*RedisSessionStore)(nil)
	_ auth.PermissionCache = (*RedisPermissionCache)(nil)
	_ lock.Provider        = (*RedisLocker)(nil)
)
