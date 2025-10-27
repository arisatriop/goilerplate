package auth

import (
	"context"
	"fmt"
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

// HasPermission checks if a user has a specific permission
//
// Cache Strategy (Simple):
// - If Redis is ENABLED: Check cache only (assumes all permissions cached on login)
// - If Redis is DISABLED: Query database normally (3 sources)
//
// Note: Call CacheAllUserPermissions() after login to populate the cache
//
// Permission sources (when Redis disabled):
// 1. User-specific permission override (user_permissions)
//   - is_granted = true: custom grant (user has permission)
//   - is_granted = false: revoked (user doesn't have permission)
//
// 2. Role-based permissions (user -> roles -> role_permissions)
// 3. Menu-based permissions (user -> roles -> role_menus -> menus + children -> menu_permissions)
func (s *PermissionService) HasPermission(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	// If Redis is enabled, ONLY check cache (write-through cache)
	if s.cacheService.IsEnabled() {
		cachedPermission, err := s.cacheService.GetCachedUserPermission(ctx, userID, permissionSlug)
		if err != nil {
			// Redis error - deny access for safety
			return false, fmt.Errorf("failed to get chaced user permission: %w", err)
		}

		// Return cached result (found = true, not found = false)
		// If not found in cache, it means user doesn't have this permission
		return cachedPermission, nil
	}

	// Redis is disabled - perform full database permission check
	return s.computePermission(ctx, userID, permissionSlug)
}

// CacheAllUserPermissions fetches and caches ALL permissions that the user has
// This includes permissions from:
// - role_permissions (via user's roles)
// - menu_permissions (via user's role menus and their children)
// - user_permissions overrides (is_granted = true OR false)
//
// Important: user_permissions can REVOKE permissions from roles/menus:
// - is_granted = true: User has this permission (even if not in roles/menus)
// - is_granted = false: User does NOT have this permission (even if in roles/menus)
//
// Call this after login to pre-populate the permission cache
func (s *PermissionService) CacheAllUserPermissions(ctx context.Context, userID string) error {

	// Step 1: Collect all permissions from roles and menus
	permissionMap := make(map[string]struct{})

	// Get permissions from role_permissions (via user's roles)
	rolePermSlugs, err := s.repo.GetRolePermissionSlugs(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get role permission slug: %w", err)
	}
	for _, slug := range rolePermSlugs {
		permissionMap[slug] = struct{}{}
	}

	// Get permissions from menu_permissions (via role menus + children)
	menuPermSlugs, err := s.getMenuPermissionSlugs(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get menu permission slugs: %w", err)
	}
	for _, slug := range menuPermSlugs {
		permissionMap[slug] = struct{}{}
	}

	// Step 2: Apply user_permissions overrides (grants AND revocations)
	userPermOverrides, err := s.repo.GetAllUserPermissionOverrides(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user permission overrides: %w", err)
	}

	for slug, isGranted := range userPermOverrides {
		if isGranted {
			// Explicitly granted - add to permissions
			permissionMap[slug] = struct{}{}
		} else {
			// Explicitly revoked - remove from permissions
			delete(permissionMap, slug)
		}
	}

	if err := s.cacheService.CacheUserPermission(ctx, userID, permissionMap, sessionDuration); err != nil {
		return fmt.Errorf("failed to cache user permissions: %w", err)
	}

	return nil
}

// computePermission computes a single permission without caching
func (s *PermissionService) computePermission(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	// Step 1: Check user_permissions (explicit grant/revoke)
	userPermission, err := s.repo.GetUserPermissionOverride(ctx, userID, permissionSlug)
	if err != nil {
		return false, err
	}

	if userPermission != nil {
		return *userPermission, nil
	}

	// Step 2: Check role_permissions
	hasRolePermission, err := s.repo.HasRolePermission(ctx, userID, permissionSlug)
	if err != nil {
		return false, err
	}

	if hasRolePermission {
		return true, nil
	}

	// Step 3: Check menu_permissions (including children)
	return s.hasMenuPermissionWithChildren(ctx, userID, permissionSlug)
}

// getMenuPermissionSlugs gets all permission slugs from menus (including children)
func (s *PermissionService) getMenuPermissionSlugs(ctx context.Context, userID string) ([]string, error) {
	// Get parent menus from user's roles
	parentMenus, err := s.repo.GetRoleMenus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role menus: %w", err)
	}

	if len(parentMenus) == 0 {
		return []string{}, nil
	}

	// Get all menu IDs (parent + children recursively)
	allMenuIDs := s.getAllMenuIDs(ctx, parentMenus)

	// Get permission slugs for these menus
	return s.repo.GetMenuPermissionSlugs(ctx, allMenuIDs)
}

// hasMenuPermissionWithChildren checks if user has permission through role menus
// This includes permissions from parent menus AND all their child menus (tree structure)
func (s *PermissionService) hasMenuPermissionWithChildren(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	// Step 1: Get parent menus from user's roles
	parentMenus, err := s.repo.GetRoleMenus(ctx, userID)
	if err != nil {
		return false, err
	}

	if len(parentMenus) == 0 {
		return false, nil
	}

	// Step 2: Get all menu IDs (parent + children recursively)
	allMenuIDs := s.getAllMenuIDs(ctx, parentMenus)

	// Step 3: Check if any of these menus have the required permission
	return s.repo.HasMenuPermissionByMenuIDs(ctx, allMenuIDs, permissionSlug)
}

// getAllMenuIDs recursively collects all menu IDs from parent menus and their children
func (s *PermissionService) getAllMenuIDs(ctx context.Context, menus []Menu) []string {
	if len(menus) == 0 {
		return []string{}
	}

	var allIDs []string

	for _, menu := range menus {
		// Add current menu ID
		allIDs = append(allIDs, menu.ID)

		// Get children from repository
		children, err := s.repo.GetMenusByParentIDs(ctx, []string{menu.ID})
		if err != nil || len(children) == 0 {
			continue
		}

		// Recursively get children IDs
		childIDs := s.getAllMenuIDs(ctx, children)
		allIDs = append(allIDs, childIDs...)
	}

	return allIDs
}
