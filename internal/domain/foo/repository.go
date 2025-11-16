package foo

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	// CreateFoo()
	// UpdateFoo()
	// DeleteFoo()
	// SoftDeleteFoo()

	// GetFooList()
	// GetFooByID()
	// CountFoos()
}
