package auth

// Menu represents a menu item in the system
type Menu struct {
	ID           string
	ParentID     *string
	Name         string
	Slug         string
	Icon         *string
	Route        *string
	DisplayOrder float64
	IsActive     bool
	Permissions  []string
	Children     []Menu
}
