package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/lock"
)

// MemorySessionStore is an in-process auth.SessionStore with a per-user index.
// It is only consistent within one instance: with several instances a revoked session
// may still be served by another instance for up to the TTL.
type MemorySessionStore struct {
	mu        sync.Mutex
	ttl       time.Duration
	entries   map[string]memorySession
	byUser    map[string]map[string]struct{}
	nextSweep time.Time
	now       func() time.Time
}

type memorySession struct {
	session   auth.UserSession
	expiresAt time.Time
}

// NewMemorySessionStore creates a MemorySessionStore that keeps sessions for at most ttl.
func NewMemorySessionStore(ttl time.Duration) *MemorySessionStore {
	return &MemorySessionStore{
		ttl:     ttl,
		entries: make(map[string]memorySession),
		byUser:  make(map[string]map[string]struct{}),
		now:     time.Now,
	}
}

func (s *MemorySessionStore) Get(_ context.Context, sessionID string) (*auth.UserSession, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[sessionID]
	if !ok {
		return nil, false, nil
	}
	if !s.now().Before(entry.expiresAt) {
		s.remove(sessionID, entry.session.UserID)
		return nil, false, nil
	}

	session := entry.session
	return &session, true, nil
}

func (s *MemorySessionStore) Set(_ context.Context, session *auth.UserSession) error {
	now := s.now()
	ttl := sessionTTL(session, s.ttl, now)
	if ttl <= 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweep(now)
	s.entries[session.ID] = memorySession{session: *session, expiresAt: now.Add(ttl)}
	if s.byUser[session.UserID] == nil {
		s.byUser[session.UserID] = make(map[string]struct{})
	}
	s.byUser[session.UserID][session.ID] = struct{}{}

	return nil
}

func (s *MemorySessionStore) Delete(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.entries[sessionID]; ok {
		s.remove(sessionID, entry.session.UserID)
	}
	return nil
}

func (s *MemorySessionStore) DeleteByUser(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID := range s.byUser[userID] {
		delete(s.entries, sessionID)
	}
	delete(s.byUser, userID)
	return nil
}

// remove deletes a session and its index entry. The caller must hold s.mu.
func (s *MemorySessionStore) remove(sessionID, userID string) {
	delete(s.entries, sessionID)
	if ids := s.byUser[userID]; ids != nil {
		delete(ids, sessionID)
		if len(ids) == 0 {
			delete(s.byUser, userID)
		}
	}
}

// sweep drops expired entries at most once per TTL so memory stays bounded.
// The caller must hold s.mu.
func (s *MemorySessionStore) sweep(now time.Time) {
	if now.Before(s.nextSweep) {
		return
	}
	for sessionID, entry := range s.entries {
		if !now.Before(entry.expiresAt) {
			s.remove(sessionID, entry.session.UserID)
		}
	}
	s.nextSweep = now.Add(s.ttl)
}

// MemoryPermissionCache is an in-process auth.PermissionCache.
// Invalidation only reaches the instance that performs it.
type MemoryPermissionCache struct {
	mu        sync.Mutex
	ttl       time.Duration
	entries   map[string]memoryPermissions
	nextSweep time.Time
	now       func() time.Time
}

type memoryPermissions struct {
	permissions []string
	expiresAt   time.Time
}

// NewMemoryPermissionCache creates a MemoryPermissionCache that keeps permissions for ttl.
func NewMemoryPermissionCache(ttl time.Duration) *MemoryPermissionCache {
	return &MemoryPermissionCache{
		ttl:     ttl,
		entries: make(map[string]memoryPermissions),
		now:     time.Now,
	}
}

func (c *MemoryPermissionCache) Get(_ context.Context, userID string) ([]string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[userID]
	if !ok {
		return nil, false, nil
	}
	if !c.now().Before(entry.expiresAt) {
		delete(c.entries, userID)
		return nil, false, nil
	}

	return append([]string(nil), entry.permissions...), true, nil
}

func (c *MemoryPermissionCache) Set(_ context.Context, userID string, permissions []string) error {
	if c.ttl <= 0 {
		return nil
	}

	now := c.now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !now.Before(c.nextSweep) {
		for id, entry := range c.entries {
			if !now.Before(entry.expiresAt) {
				delete(c.entries, id)
			}
		}
		c.nextSweep = now.Add(c.ttl)
	}

	c.entries[userID] = memoryPermissions{
		permissions: append([]string{}, permissions...),
		expiresAt:   now.Add(c.ttl),
	}
	return nil
}

func (c *MemoryPermissionCache) Invalidate(_ context.Context, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, userID)
	return nil
}

func (c *MemoryPermissionCache) InvalidateAll(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]memoryPermissions)
	return nil
}

// MemoryLocker is a lock.Provider scoped to one process.
type MemoryLocker struct {
	mu     sync.Mutex
	locks  map[string]memoryLock
	tokens atomic.Uint64
	now    func() time.Time
}

type memoryLock struct {
	token     uint64
	expiresAt time.Time
}

// NewMemoryLocker creates a MemoryLocker.
func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{
		locks: make(map[string]memoryLock),
		now:   time.Now,
	}
}

func (l *MemoryLocker) TryLock(_ context.Context, key string, ttl time.Duration) (lock.Release, bool, error) {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if held, ok := l.locks[key]; ok && now.Before(held.expiresAt) {
		return nil, false, nil
	}

	token := l.tokens.Add(1)
	l.locks[key] = memoryLock{token: token, expiresAt: now.Add(ttl)}

	release := func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()

		if held, ok := l.locks[key]; ok && held.token == token {
			delete(l.locks, key)
		}
		return nil
	}
	return release, true, nil
}

// sessionTTL caps the cache TTL at the session's own remaining lifetime.
func sessionTTL(session *auth.UserSession, cacheTTL time.Duration, now time.Time) time.Duration {
	ttl := cacheTTL
	if remaining := session.ExpiresAt.Sub(now); remaining < ttl {
		ttl = remaining
	}
	return ttl
}

var (
	_ auth.SessionStore    = (*MemorySessionStore)(nil)
	_ auth.PermissionCache = (*MemoryPermissionCache)(nil)
	_ lock.Provider        = (*MemoryLocker)(nil)
)
