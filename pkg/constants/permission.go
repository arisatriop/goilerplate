package constants

// Permission constants using resource.action format
// These should match the permission slugs in the database

// Foo Resource Permissions
const (
	PermissionFooList   = "foo.list"
	PermissionFooGet    = "foo.get"
	PermissionFooCreate = "foo.create"
	PermissionFooUpdate = "foo.update"
	PermissionFooDelete = "foo.delete"
)

// Bar Resource Permissions
const (
	PermissionBarList   = "bar.list"
	PermissionBarGet    = "bar.get"
	PermissionBarCreate = "bar.create"
	PermissionBarUpdate = "bar.update"
	PermissionBarDelete = "bar.delete"
)

// Add more resource permissions here as needed
