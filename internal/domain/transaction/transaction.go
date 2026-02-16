package transaction

import "context"

// Transaction represents a unit of work abstraction
// This interface allows the application layer to manage transactions
// without depending on specific infrastructure (GORM, SQL, etc.)
type Transaction interface {
	// Do executes the given function within a transaction context
	// If the function returns an error, the transaction is rolled back
	// Otherwise, the transaction is committed
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
