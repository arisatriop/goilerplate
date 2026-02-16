package order

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	GetOrderByID(ctx context.Context, orderID string, storeID string) (*Order, error)
	GetListOrder(ctx context.Context, filter *Filter) ([]*Order, error)
	CountOrders(ctx context.Context, filter *Filter) (int64, error)
}
