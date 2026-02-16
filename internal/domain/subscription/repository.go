package subscription

import (
	"context"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateSubscription(ctx context.Context, subscription *Subscription) (*Subscription, error)
}
