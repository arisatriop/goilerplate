package auth

import (
	"context"
	"sort"
)

// MenuService handles menu-related operations
type MenuService struct {
	repo Repository
}

// NewMenuService creates a new menu service
func NewMenuService(repo Repository) *MenuService {
	return &MenuService{
		repo: repo,
	}
}

// BuildMenuTree builds a hierarchical menu tree from parent menus, fetching children on-demand
func (s *MenuService) BuildMenuTree(ctx context.Context, parentMenus []Menu) []Menu {
	if len(parentMenus) == 0 {
		return []Menu{}
	}

	// Sort parent menus by display_order
	sort.Slice(parentMenus, func(i, j int) bool {
		return parentMenus[i].DisplayOrder < parentMenus[j].DisplayOrder
	})

	// Build tree by fetching children for each parent
	result := make([]Menu, len(parentMenus))
	for i, parent := range parentMenus {
		result[i] = s.buildMenuWithChildren(ctx, parent)
	}

	return result
}

// buildMenuWithChildren recursively builds a menu with all its children by fetching from repository
func (s *MenuService) buildMenuWithChildren(ctx context.Context, menu Menu) Menu {
	// Create a copy of the current menu
	menuCopy := Menu{
		ID:           menu.ID,
		ParentID:     menu.ParentID,
		Name:         menu.Name,
		Slug:         menu.Slug,
		Icon:         menu.Icon,
		Route:        menu.Route,
		DisplayOrder: menu.DisplayOrder,
		IsActive:     menu.IsActive,
		Children:     []Menu{},
	}

	// Fetch children from repository
	children, err := s.repo.GetMenusByParentIDs(ctx, []string{menu.ID})
	if err != nil || len(children) == 0 {
		return menuCopy
	}

	// Sort children by display_order
	sort.Slice(children, func(i, j int) bool {
		return children[i].DisplayOrder < children[j].DisplayOrder
	})

	// Recursively build each child with its descendants
	menuCopy.Children = make([]Menu, len(children))
	for i, child := range children {
		menuCopy.Children[i] = s.buildMenuWithChildren(ctx, child)
	}

	return menuCopy
}
