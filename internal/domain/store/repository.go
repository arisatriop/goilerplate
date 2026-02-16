package store

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateStore(ctx context.Context, store *Store) (*Store, error)
	GetStoreByUserID(ctx context.Context, id uuid.UUID) (*Store, error)
	GetStoreByID(ctx context.Context, id uuid.UUID) (*Store, error)
}
