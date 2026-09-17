package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePermissionCache is a map-backed PermissionCache that can simulate cache failures.
type fakePermissionCache struct {
	entries map[string][]string
	err     error
}

func newFakePermissionCache() *fakePermissionCache {
	return &fakePermissionCache{entries: make(map[string][]string)}
}

func (c *fakePermissionCache) Get(_ context.Context, userID string) ([]string, bool, error) {
	if c.err != nil {
		return nil, false, c.err
	}
	permissions, ok := c.entries[userID]
	return permissions, ok, nil
}

func (c *fakePermissionCache) Set(_ context.Context, userID string, permissions []string) error {
	if c.err != nil {
		return c.err
	}
	c.entries[userID] = permissions
	return nil
}

func (c *fakePermissionCache) Invalidate(_ context.Context, userID string) error {
	if c.err != nil {
		return c.err
	}
	delete(c.entries, userID)
	return nil
}

func (c *fakePermissionCache) InvalidateAll(context.Context) error {
	if c.err != nil {
		return c.err
	}
	c.entries = make(map[string][]string)
	return nil
}

func TestPermissionService_GetUserFinalPermissions_CachesMergedResult(t *testing.T) {
	// Arrange
	repo := &stubRepository{
		rolePermissions: []string{"foo.list", "foo.delete"},
		overrides:       map[string]bool{"foo.delete": false, "bar.list": true},
	}
	cache := newFakePermissionCache()
	service := NewPermissionService(repo, cache)
	ctx := context.Background()

	// Act
	first, err := service.GetUserFinalPermissions(ctx, "u1")
	require.NoError(t, err)
	second, err := service.GetUserFinalPermissions(ctx, "u1")
	require.NoError(t, err)

	// Assert
	assert.ElementsMatch(t, []string{"foo.list", "bar.list"}, first)
	assert.ElementsMatch(t, first, second)
	assert.Equal(t, 1, repo.permissionReads, "second call is served from cache")
}

func TestPermissionService_InvalidateUserPermissions_NextCheckSeesChange(t *testing.T) {
	// Arrange
	repo := &stubRepository{rolePermissions: []string{"foo.delete"}}
	service := NewPermissionService(repo, newFakePermissionCache())
	ctx := context.Background()

	allowed, err := service.HasPermission(ctx, "u1", "foo.delete")
	require.NoError(t, err)
	require.True(t, allowed)

	// Act: the permission is revoked and the cache invalidated
	repo.rolePermissions = nil
	require.NoError(t, service.InvalidateUserPermissions(ctx, "u1"))

	// Assert
	allowed, err = service.HasPermission(ctx, "u1", "foo.delete")
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestPermissionService_InvalidateAllPermissions(t *testing.T) {
	repo := &stubRepository{rolePermissions: []string{"foo.list"}}
	cache := newFakePermissionCache()
	service := NewPermissionService(repo, cache)
	ctx := context.Background()
	_, err := service.GetUserFinalPermissions(ctx, "u1")
	require.NoError(t, err)
	_, err = service.GetUserFinalPermissions(ctx, "u2")
	require.NoError(t, err)

	require.NoError(t, service.InvalidateAllPermissions(ctx))

	assert.Empty(t, cache.entries)
}

func TestPermissionService_CacheFailureFallsBackToRepository(t *testing.T) {
	// Arrange
	repo := &stubRepository{rolePermissions: []string{"foo.list"}}
	cache := newFakePermissionCache()
	cache.err = errors.New("redis down")
	service := NewPermissionService(repo, cache)

	// Act
	allowed, err := service.HasPermission(context.Background(), "u1", "foo.list")

	// Assert
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Error(t, service.InvalidateUserPermissions(context.Background(), "u1"), "invalidation failures are reported")
}
