package auth

import "context"

// SessionStore caches sessions by ID in front of the repository.
// The repository stays the source of truth: a miss is never an authorization decision.
// Implementations (noop, memory, redis) live in internal/infrastructure/cache and are
// selected at wiring time.
type SessionStore interface {
	// Get returns the cached session and true, or nil and false on a cache miss.
	Get(ctx context.Context, sessionID string) (*UserSession, bool, error)
	// Set caches a session (active or not) for at most the configured cache TTL.
	Set(ctx context.Context, session *UserSession) error
	// Delete evicts one session.
	Delete(ctx context.Context, sessionID string) error
	// DeleteByUser evicts every cached session of a user without scanning the keyspace.
	DeleteByUser(ctx context.Context, userID string) error
}

// PermissionCache caches a user's final (merged) permission slugs.
type PermissionCache interface {
	// Get returns the cached permissions and true, or nil and false on a cache miss.
	Get(ctx context.Context, userID string) ([]string, bool, error)
	// Set caches permissions for the configured TTL.
	Set(ctx context.Context, userID string, permissions []string) error
	// Invalidate evicts one user's permissions (user roles or overrides changed).
	Invalidate(ctx context.Context, userID string) error
	// InvalidateAll evicts every user's permissions (role permissions or role menus changed).
	InvalidateAll(ctx context.Context) error
}
