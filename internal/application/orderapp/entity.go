package orderapp

import (
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/orderitem"
)

type Order struct {
	Order      *order.Order
	OrderItems []orderitem.OrderItem
}
