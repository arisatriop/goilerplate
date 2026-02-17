package auth

import (
	"context"
)

// PermissionService handles permission-related operations
type PermissionService struct {
	repo         Repository
	cacheService *CacheService
}

// NewPermissionService creates a new permission service
func NewPermissionService(repo Repository, cacheService *CacheService) *PermissionService {
	return &PermissionService{
		repo:         repo,
		cacheService: cacheService,
	}
}

// GetUserFinalPermissions gets merged user permissions (role permissions + user overrides)
func (s *PermissionService) GetUserFinalPermissions(ctx context.Context, userID string) ([]string, error) {
	// Try to get from cache first
	if s.cacheService.IsEnabled() {
		if cachedPermissions, found, err := s.cacheService.GetCachedUserPermissions(ctx, userID); err == nil && found {
			return cachedPermissions, nil
		}
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

	// Merge permissions
	return s.mergePermissions(rolePermissions, userPermissionOverrides), nil
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

// CacheAllUserPermissions caches all user permissions (merged role + user overrides) to Redis
func (s *PermissionService) CacheAllUserPermissions(ctx context.Context, userID string) error {
	if !s.cacheService.IsEnabled() {
		return nil // Skip if Redis is disabled
	}

	// Get final merged permissions
	finalPermissions, err := s.GetUserFinalPermissions(ctx, userID)
	if err != nil {
		return err
	}

	// Convert to map for faster lookup
	permissionMap := make(map[string]struct{})
	for _, permission := range finalPermissions {
		permissionMap[permission] = struct{}{}
	}

	// Cache with session duration TTL to match user session lifetime
	ttl := SessionDuration
	return s.cacheService.CacheUserPermissions(ctx, userID, permissionMap, ttl)
}

// InvalidateUserPermissions clears cached permissions for a specific user
// This should be called when user's roles or permissions are modified
func (s *PermissionService) InvalidateUserPermissions(ctx context.Context, userID string) error {
	if !s.cacheService.IsEnabled() {
		return nil // Skip if Redis is disabled
	}

	return s.cacheService.InvalidateUserPermissions(ctx, userID)
}

// InvalidateAllPermissions clears cached permissions for all users
// This should be called when roles, permissions, or menus are modified globally
func (s *PermissionService) InvalidateAllPermissions(ctx context.Context) error {
	if !s.cacheService.IsEnabled() {
		return nil // Skip if Redis is disabled
	}

	return s.cacheService.InvalidateAllPermissions(ctx)
}
