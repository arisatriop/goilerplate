package orderstatushistory

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateOrderStatusHistory(ctx context.Context, history *OrderStatusHistory) error
}
