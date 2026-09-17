// Package lock defines a best-effort mutual exclusion primitive for work that must not
// run concurrently, such as background jobs or duplicate requests.
package lock

import (
	"context"
	"time"
)

// Release frees a held lock. Releasing a lock that already expired or was taken over
// by another holder is a no-op.
type Release func(ctx context.Context) error

// Provider acquires named locks. Implementations (memory, redis) live in
// internal/infrastructure/cache: memory locks are per process, redis locks are shared
// by every instance.
type Provider interface {
	// TryLock acquires key for at most ttl without waiting.
	// It returns acquired=false (and a nil Release) when another holder owns the key.
	TryLock(ctx context.Context, key string, ttl time.Duration) (release Release, acquired bool, err error)
}
