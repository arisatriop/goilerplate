package orderitem

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	CreateOrderItem(ctx context.Context, orderItem *OrderItem) error
	GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]*OrderItem, error)
}
