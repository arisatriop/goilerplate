package cache

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/lock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTTL = 150 * time.Millisecond

// newTestRedis connects to REDIS_TEST_ADDR (database REDIS_TEST_DB, default 15) and
// returns a unique key prefix whose keys are removed after the test.
// Redis tests are skipped when REDIS_TEST_ADDR is not set.
func newTestRedis(t *testing.T) (*redis.Client, string) {
	t.Helper()

	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set; skipping Redis integration test")
	}
	db := 15
	if value := os.Getenv("REDIS_TEST_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		require.NoError(t, err)
		db = parsed
	}

	client := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	require.NoError(t, client.Ping(context.Background()).Err())

	token, err := randomToken()
	require.NoError(t, err)
	prefix := "test:" + token + ":"

	t.Cleanup(func() {
		ctx := context.Background()
		iter := client.Scan(ctx, 0, prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			_ = client.Del(ctx, iter.Val()).Err()
		}
		_ = client.Close()
	})

	return client, prefix
}

type sessionStoreFactory func(t *testing.T, ttl time.Duration) auth.SessionStore

func sessionStoreFactories() map[string]sessionStoreFactory {
	return map[string]sessionStoreFactory{
		"memory": func(_ *testing.T, ttl time.Duration) auth.SessionStore {
			return NewMemorySessionStore(ttl)
		},
		"redis": func(t *testing.T, ttl time.Duration) auth.SessionStore {
			client, prefix := newTestRedis(t)
			store := NewRedisSessionStore(client, ttl)
			store.prefix = prefix
			return store
		},
	}
}

func newSession(id, userID string) *auth.UserSession {
	return &auth.UserSession{
		ID:        id,
		UserID:    userID,
		DeviceID:  "device-" + id,
		IsActive:  true,
		ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond),
	}
}

func TestSessionStore_Contract(t *testing.T) {
	for name, factory := range sessionStoreFactories() {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			t.Run("miss then hit", func(t *testing.T) {
				store := factory(t, time.Minute)
				session := newSession("s1", "u1")

				_, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				assert.False(t, found)

				require.NoError(t, store.Set(ctx, session))
				got, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, session.UserID, got.UserID)
				assert.Equal(t, session.DeviceID, got.DeviceID)
				assert.True(t, got.IsActive)
				assert.True(t, session.ExpiresAt.Equal(got.ExpiresAt))
			})

			t.Run("inactive sessions are cached too", func(t *testing.T) {
				store := factory(t, time.Minute)
				session := newSession("s1", "u1")
				session.IsActive = false

				require.NoError(t, store.Set(ctx, session))
				got, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				require.True(t, found)
				assert.False(t, got.IsActive)
			})

			t.Run("entries expire after ttl", func(t *testing.T) {
				store := factory(t, testTTL)
				require.NoError(t, store.Set(ctx, newSession("s1", "u1")))

				time.Sleep(testTTL + 100*time.Millisecond)

				_, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				assert.False(t, found)
			})

			t.Run("ttl is capped by session expiry", func(t *testing.T) {
				store := factory(t, time.Minute)
				session := newSession("s1", "u1")
				session.ExpiresAt = time.Now().Add(testTTL)
				require.NoError(t, store.Set(ctx, session))

				time.Sleep(testTTL + 100*time.Millisecond)

				_, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				assert.False(t, found)
			})

			t.Run("expired session is not cached", func(t *testing.T) {
				store := factory(t, time.Minute)
				session := newSession("s1", "u1")
				session.ExpiresAt = time.Now().Add(-time.Second)

				require.NoError(t, store.Set(ctx, session))
				_, found, err := store.Get(ctx, "s1")
				require.NoError(t, err)
				assert.False(t, found)
			})

			t.Run("delete evicts one session", func(t *testing.T) {
				store := factory(t, time.Minute)
				require.NoError(t, store.Set(ctx, newSession("s1", "u1")))
				require.NoError(t, store.Set(ctx, newSession("s2", "u1")))

				require.NoError(t, store.Delete(ctx, "s1"))
				require.NoError(t, store.Delete(ctx, "missing"))

				_, found, _ := store.Get(ctx, "s1")
				assert.False(t, found)
				_, found, _ = store.Get(ctx, "s2")
				assert.True(t, found)
			})

			t.Run("delete by user leaves other users untouched", func(t *testing.T) {
				store := factory(t, time.Minute)
				require.NoError(t, store.Set(ctx, newSession("a1", "alice")))
				require.NoError(t, store.Set(ctx, newSession("a2", "alice")))
				require.NoError(t, store.Set(ctx, newSession("b1", "bob")))

				require.NoError(t, store.DeleteByUser(ctx, "alice"))
				require.NoError(t, store.DeleteByUser(ctx, "nobody"))

				_, found, _ := store.Get(ctx, "a1")
				assert.False(t, found)
				_, found, _ = store.Get(ctx, "a2")
				assert.False(t, found)
				_, found, _ = store.Get(ctx, "b1")
				assert.True(t, found)

				// Sessions cached after DeleteByUser are indexed again.
				require.NoError(t, store.Set(ctx, newSession("a3", "alice")))
				require.NoError(t, store.DeleteByUser(ctx, "alice"))
				_, found, _ = store.Get(ctx, "a3")
				assert.False(t, found)
			})
		})
	}
}

type permissionCacheFactory func(t *testing.T, ttl time.Duration) auth.PermissionCache

func permissionCacheFactories() map[string]permissionCacheFactory {
	return map[string]permissionCacheFactory{
		"memory": func(_ *testing.T, ttl time.Duration) auth.PermissionCache {
			return NewMemoryPermissionCache(ttl)
		},
		"redis": func(t *testing.T, ttl time.Duration) auth.PermissionCache {
			client, prefix := newTestRedis(t)
			cache := NewRedisPermissionCache(client, ttl)
			cache.prefix = prefix
			return cache
		},
	}
}

func TestPermissionCache_Contract(t *testing.T) {
	for name, factory := range permissionCacheFactories() {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			t.Run("miss then hit", func(t *testing.T) {
				cache := factory(t, time.Minute)

				_, found, err := cache.Get(ctx, "u1")
				require.NoError(t, err)
				assert.False(t, found)

				require.NoError(t, cache.Set(ctx, "u1", []string{"foo.create", "foo.list"}))
				got, found, err := cache.Get(ctx, "u1")
				require.NoError(t, err)
				assert.True(t, found)
				assert.ElementsMatch(t, []string{"foo.create", "foo.list"}, got)
			})

			t.Run("empty permission list is a hit", func(t *testing.T) {
				cache := factory(t, time.Minute)

				require.NoError(t, cache.Set(ctx, "u1", nil))
				got, found, err := cache.Get(ctx, "u1")
				require.NoError(t, err)
				assert.True(t, found)
				assert.Empty(t, got)
			})

			t.Run("entries expire after ttl", func(t *testing.T) {
				cache := factory(t, testTTL)
				require.NoError(t, cache.Set(ctx, "u1", []string{"foo.list"}))

				time.Sleep(testTTL + 100*time.Millisecond)

				_, found, err := cache.Get(ctx, "u1")
				require.NoError(t, err)
				assert.False(t, found)
			})

			t.Run("invalidate one user", func(t *testing.T) {
				cache := factory(t, time.Minute)
				require.NoError(t, cache.Set(ctx, "u1", []string{"foo.list"}))
				require.NoError(t, cache.Set(ctx, "u2", []string{"foo.list"}))

				require.NoError(t, cache.Invalidate(ctx, "u1"))

				_, found, _ := cache.Get(ctx, "u1")
				assert.False(t, found)
				_, found, _ = cache.Get(ctx, "u2")
				assert.True(t, found)
			})

			t.Run("invalidate all users", func(t *testing.T) {
				cache := factory(t, time.Minute)
				for i := 0; i < 3; i++ {
					require.NoError(t, cache.Set(ctx, "u"+strconv.Itoa(i), []string{"foo.list"}))
				}

				require.NoError(t, cache.InvalidateAll(ctx))

				for i := 0; i < 3; i++ {
					_, found, _ := cache.Get(ctx, "u"+strconv.Itoa(i))
					assert.False(t, found)
				}

				require.NoError(t, cache.Set(ctx, "u9", []string{"foo.list"}))
				_, found, _ := cache.Get(ctx, "u9")
				assert.True(t, found, "cache keeps working after InvalidateAll")
			})
		})
	}
}

func TestRedisPermissionCache_InvalidateAllManyUsers(t *testing.T) {
	// Arrange: more users than one delete batch
	client, prefix := newTestRedis(t)
	cache := NewRedisPermissionCache(client, time.Minute)
	cache.prefix = prefix
	ctx := context.Background()
	users := redisDeleteBatchSize + 7
	for i := 0; i < users; i++ {
		require.NoError(t, cache.Set(ctx, "u"+strconv.Itoa(i), []string{"p"}))
	}

	// Act
	require.NoError(t, cache.InvalidateAll(ctx))

	// Assert
	remaining, err := client.Keys(ctx, prefix+"*").Result()
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

type lockerFactory func(t *testing.T) lock.Provider

func lockerFactories() map[string]lockerFactory {
	return map[string]lockerFactory{
		"memory": func(*testing.T) lock.Provider { return NewMemoryLocker() },
		"redis": func(t *testing.T) lock.Provider {
			client, prefix := newTestRedis(t)
			locker := NewRedisLocker(client)
			locker.prefix = prefix
			return locker
		},
	}
}

func TestLockProvider_Contract(t *testing.T) {
	for name, factory := range lockerFactories() {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			t.Run("exclusive until released", func(t *testing.T) {
				locker := factory(t)

				release, acquired, err := locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				require.True(t, acquired)

				_, acquired, err = locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				assert.False(t, acquired)

				_, acquired, err = locker.TryLock(ctx, "other-job", time.Minute)
				require.NoError(t, err)
				assert.True(t, acquired, "different keys do not block each other")

				require.NoError(t, release(ctx))
				_, acquired, err = locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				assert.True(t, acquired)
			})

			t.Run("expires after ttl", func(t *testing.T) {
				locker := factory(t)

				_, acquired, err := locker.TryLock(ctx, "job", testTTL)
				require.NoError(t, err)
				require.True(t, acquired)

				time.Sleep(testTTL + 100*time.Millisecond)

				_, acquired, err = locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				assert.True(t, acquired)
			})

			t.Run("stale release does not free a new holder", func(t *testing.T) {
				locker := factory(t)

				staleRelease, acquired, err := locker.TryLock(ctx, "job", testTTL)
				require.NoError(t, err)
				require.True(t, acquired)
				time.Sleep(testTTL + 100*time.Millisecond)

				_, acquired, err = locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				require.True(t, acquired)

				require.NoError(t, staleRelease(ctx))

				_, acquired, err = locker.TryLock(ctx, "job", time.Minute)
				require.NoError(t, err)
				assert.False(t, acquired, "the new holder still owns the lock")
			})

			t.Run("only one concurrent caller acquires", func(t *testing.T) {
				locker := factory(t)
				const callers = 20

				var wg sync.WaitGroup
				var mu sync.Mutex
				winners := 0
				for i := 0; i < callers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, acquired, err := locker.TryLock(ctx, "race", time.Minute)
						assert.NoError(t, err)
						if acquired {
							mu.Lock()
							winners++
							mu.Unlock()
						}
					}()
				}
				wg.Wait()

				assert.Equal(t, 1, winners)
			})
		})
	}
}

func TestNoopStores_AlwaysMiss(t *testing.T) {
	ctx := context.Background()

	sessions := NoopSessionStore{}
	require.NoError(t, sessions.Set(ctx, newSession("s1", "u1")))
	_, found, err := sessions.Get(ctx, "s1")
	require.NoError(t, err)
	assert.False(t, found)
	assert.NoError(t, sessions.Delete(ctx, "s1"))
	assert.NoError(t, sessions.DeleteByUser(ctx, "u1"))

	permissions := NoopPermissionCache{}
	require.NoError(t, permissions.Set(ctx, "u1", []string{"p"}))
	_, found, err = permissions.Get(ctx, "u1")
	require.NoError(t, err)
	assert.False(t, found)
	assert.NoError(t, permissions.Invalidate(ctx, "u1"))
	assert.NoError(t, permissions.InvalidateAll(ctx))
}

func TestMemorySessionStore_SweepsExpiredEntries(t *testing.T) {
	// Arrange: a controllable clock
	now := time.Now()
	store := NewMemorySessionStore(time.Minute)
	store.now = func() time.Time { return now }
	ctx := context.Background()
	require.NoError(t, store.Set(ctx, newSession("old", "u1")))

	// Act: a later write after the TTL triggers a sweep
	now = now.Add(2 * time.Minute)
	require.NoError(t, store.Set(ctx, newSession("new", "u2")))

	// Assert
	store.mu.Lock()
	defer store.mu.Unlock()
	assert.NotContains(t, store.entries, "old")
	assert.NotContains(t, store.byUser, "u1")
	assert.Contains(t, store.entries, "new")
}
