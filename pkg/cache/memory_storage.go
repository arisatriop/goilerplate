package cache

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// memorySweepInterval bounds how often expired entries are purged on writes.
const memorySweepInterval = time.Minute

// MemoryStorage implements fiber.Storage in process memory. Entries are not shared
// between instances or kept across restarts; use FiberStorage (Redis) for that.
type MemoryStorage struct {
	mu        sync.Mutex
	entries   map[string]memoryEntry
	nextSweep time.Time
	now       func() time.Time
}

type memoryEntry struct {
	value     []byte
	expiresAt time.Time // zero means no expiry
}

// NewMemoryStorage returns an empty in-memory fiber.Storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		entries: make(map[string]memoryEntry),
		now:     time.Now,
	}
}

// Get returns a copy of the value, or nil when the key is missing or expired.
func (s *MemoryStorage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[key]
	if !ok {
		return nil, nil
	}
	if entry.expired(s.now()) {
		delete(s.entries, key)
		return nil, nil
	}
	return append([]byte(nil), entry.value...), nil
}

// Set stores a copy of val. An exp of 0 keeps the entry until it is deleted.
func (s *MemoryStorage) Set(key string, val []byte, exp time.Duration) error {
	now := s.now()
	entry := memoryEntry{value: append([]byte(nil), val...)}
	if exp > 0 {
		entry.expiresAt = now.Add(exp)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !now.Before(s.nextSweep) {
		for k, e := range s.entries {
			if e.expired(now) {
				delete(s.entries, k)
			}
		}
		s.nextSweep = now.Add(memorySweepInterval)
	}

	s.entries[key] = entry
	return nil
}

func (s *MemoryStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

func (s *MemoryStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]memoryEntry)
	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (e memoryEntry) expired(now time.Time) bool {
	return !e.expiresAt.IsZero() && !now.Before(e.expiresAt)
}

var _ fiber.Storage = (*MemoryStorage)(nil)
