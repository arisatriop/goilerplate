package cache

import (
	"context"

	"goilerplate/internal/domain/auth"
)

// NoopSessionStore never caches: every lookup is a miss, so sessions are always read
// from the repository.
type NoopSessionStore struct{}

func (NoopSessionStore) Get(context.Context, string) (*auth.UserSession, bool, error) {
	return nil, false, nil
}

func (NoopSessionStore) Set(context.Context, *auth.UserSession) error { return nil }

func (NoopSessionStore) Delete(context.Context, string) error { return nil }

func (NoopSessionStore) DeleteByUser(context.Context, string) error { return nil }

// NoopPermissionCache never caches: permissions are always read from the repository.
type NoopPermissionCache struct{}

func (NoopPermissionCache) Get(context.Context, string) ([]string, bool, error) {
	return nil, false, nil
}

func (NoopPermissionCache) Set(context.Context, string, []string) error { return nil }

func (NoopPermissionCache) Invalidate(context.Context, string) error { return nil }

func (NoopPermissionCache) InvalidateAll(context.Context) error { return nil }

var (
	_ auth.SessionStore    = NoopSessionStore{}
	_ auth.PermissionCache = NoopPermissionCache{}
)
