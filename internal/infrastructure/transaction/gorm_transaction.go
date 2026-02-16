package transaction

import (
	"context"
	"goilerplate/internal/domain/transaction"

	"gorm.io/gorm"
)

// TxKey is the context key for storing transaction DB
type TxKey struct{}

// gormTransaction implements the Transaction interface using GORM
type gormTransaction struct {
	db *gorm.DB
}

// NewGormTransaction creates a new GORM-based transaction manager
func NewGormTransaction(db *gorm.DB) transaction.Transaction {
	return &gormTransaction{
		db: db,
	}
}

// Do executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (t *gormTransaction) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// Use GORM's built-in Transaction method
	// This provides automatic begin, commit, and rollback
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store transaction DB in context
		txCtx := context.WithValue(ctx, TxKey{}, tx)
		return fn(txCtx)
	})
}

// GetTxFromContext retrieves transaction DB from context
// Returns nil if no transaction exists
func GetTxFromContext(ctx context.Context) *gorm.DB {
	if tx := ctx.Value(TxKey{}); tx != nil {
		if gormTx, ok := tx.(*gorm.DB); ok {
			return gormTx
		}
	}
	return nil
}
