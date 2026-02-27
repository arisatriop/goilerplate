package constants

// Permission constants using resource.action format
// These should match the permission slugs in the database

// Template Resource Permissions
const (
	PermissionTemplateList   = "template.list"
	PermissionTemplateDetail = "template.detail"
	PermissionTemplateCreate = "template.create"
	PermissionTemplateUpdate = "template.update"
	PermissionTemplateDelete = "template.delete"
)

// Example Resource Permissions
const (
	PermissionExampleList   = "example.list"
	PermissionExampleDetail = "example.detail"
	PermissionExampleCreate = "example.create"
	PermissionExampleUpdate = "example.update"
	PermissionExampleDelete = "example.delete"
)

// Add more resource permissions here as needed
