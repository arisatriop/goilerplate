package dtoresponse

// MenuResponse represents a menu item in the response
type MenuResponse struct {
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	Icon         *string        `json:"icon"`
	Route        *string        `json:"route"`
	DisplayOrder float64        `json:"displayOrder"`
	IsActive     bool           `json:"isActive"`
	Permissions  []string       `json:"permissions"`
	Child        []MenuResponse `json:"child"`
}
