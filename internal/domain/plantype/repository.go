package plantype

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	GetStoreActiveSubsciption(ctx context.Context, storeID uuid.UUID) ([]PlanType, error)
	GetListWithPlans(ctx context.Context) ([]PlanTypeWithPlans, error)
}
