package bar

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	// CreateBar()
	// UpdateBar()
	// DeleteBar()
	// SoftDeleteBar()

	// GetBarList()
	// GetBarByID()
	// CountBars()
}
