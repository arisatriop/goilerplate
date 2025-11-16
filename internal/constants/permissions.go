package constants

// Permission constants using resource.action format
// These should match the permission slugs in the database

// Foo Resource Permissions
const (
	PermissionFooList   = "foo.list"
	PermissionFooDetail = "foo.detail"
	PermissionFooCreate = "foo.create"
	PermissionFooUpdate = "foo.update"
	PermissionFooDelete = "foo.delete"
)
