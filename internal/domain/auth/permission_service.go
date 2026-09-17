package auth

import (
	"context"
	"fmt"

	"goilerplate/pkg/logger"
)

// PermissionService handles permission-related operations
type PermissionService struct {
	repo  Repository
	cache PermissionCache
}

// NewPermissionService creates a new permission service
func NewPermissionService(repo Repository, cache PermissionCache) *PermissionService {
	return &PermissionService{
		repo:  repo,
		cache: cache,
	}
}

// GetUserFinalPermissions gets merged user permissions (role permissions + user overrides).
// Results are cached; cache failures are logged and fall back to the repository.
func (s *PermissionService) GetUserFinalPermissions(ctx context.Context, userID string) ([]string, error) {
	cachedPermissions, found, err := s.cache.Get(ctx, userID)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("reading permission cache: %w", err))
	} else if found {
		return cachedPermissions, nil
	}

	// Get user roles
	userRoles, err := s.repo.GetUserRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get permissions for user roles
	rolePermissions, err := s.repo.GetRolePermissionsByRoleIDs(ctx, userRoles)
	if err != nil {
		return nil, err
	}

	// Get user permission overrides
	userPermissionOverrides, err := s.repo.GetUserPermissionOverrides(ctx, userID)
	if err != nil {
		return nil, err
	}

	finalPermissions := s.mergePermissions(rolePermissions, userPermissionOverrides)
	if err := s.cache.Set(ctx, userID, finalPermissions); err != nil {
		logger.Error(ctx, fmt.Errorf("caching permissions: %w", err))
	}

	return finalPermissions, nil
}

// HasPermission checks if user has a specific permission after merging role and user permissions
func (s *PermissionService) HasPermission(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	finalPermissions, err := s.GetUserFinalPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if permission exists in final list
	for _, permission := range finalPermissions {
		if permission == permissionSlug {
			return true, nil
		}
	}

	return false, nil
}

// mergePermissions merges role permissions with user permission overrides
// Returns final permission list after applying user-specific grants and revocations
func (s *PermissionService) mergePermissions(rolePermissions []string, userOverrides map[string]bool) []string {
	// Start with role permissions as a set for efficient lookup
	permissionSet := make(map[string]bool)
	for _, permission := range rolePermissions {
		permissionSet[permission] = true
	}

	// Apply user permission overrides
	for permission, isGranted := range userOverrides {
		if isGranted {
			// Grant permission (add to set)
			permissionSet[permission] = true
		} else {
			// Revoke permission (remove from set)
			delete(permissionSet, permission)
		}
	}

	// Convert set back to slice
	finalPermissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		finalPermissions = append(finalPermissions, permission)
	}

	return finalPermissions
}

// InvalidateUserPermissions clears cached permissions for a specific user.
// Call it whenever the user's roles or permission overrides change.
func (s *PermissionService) InvalidateUserPermissions(ctx context.Context, userID string) error {
	if err := s.cache.Invalidate(ctx, userID); err != nil {
		return fmt.Errorf("invalidating user permissions: %w", err)
	}
	return nil
}

// InvalidateAllPermissions clears cached permissions for all users.
// Call it whenever role permissions or role menus change.
func (s *PermissionService) InvalidateAllPermissions(ctx context.Context) error {
	if err := s.cache.InvalidateAll(ctx); err != nil {
		return fmt.Errorf("invalidating all permissions: %w", err)
	}
	return nil
}
