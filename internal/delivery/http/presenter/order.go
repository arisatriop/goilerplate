package presenter

import (
	"goilerplate/internal/application/orderapp"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/orderitem"
)

func ToOrderCreateResponse(order *order.Order, orderItems []orderitem.OrderItem) dtoresponse.OrderResponse {

	var items []dtoresponse.OrderItem
	for _, item := range orderItems {
		items = append(items, dtoresponse.OrderItem{
			ProductName: item.Name,
			Note:        item.Note,
			Quantity:    item.Qty,
			Price:       item.Price.StringFixed(2),
			SubTotal:    item.SubTotal.StringFixed(2),
		})
	}

	return dtoresponse.OrderResponse{
		OrderID:       order.OrderNumber,
		OrderNumber:   order.QueueNumber,
		TableNumber:   order.TableNumber,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
		OrderType:     order.OrderType,
		GrandTotal:    order.Amount.StringFixed(2),
		Items:         items,
	}
}

func ToOrderSummaryResponse(order *order.Order) *dtoresponse.OrderSummaryResponse {
	return &dtoresponse.OrderSummaryResponse{
		ID:            order.ID,
		OrderID:       order.OrderNumber,
		OrderNumber:   order.QueueNumber,
		TableNumber:   order.TableNumber,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
		OrderType:     order.OrderType,
		OrderStatus:   order.OrderStatus,
		Amount:        order.Amount.StringFixed(2),
		CreatedAt:     order.CreatedAt,
	}
}

func ToOrderSummaryListResponse(orders []*order.Order) []*dtoresponse.OrderSummaryResponse {
	responses := make([]*dtoresponse.OrderSummaryResponse, len(orders))
	for i, order := range orders {
		responses[i] = ToOrderSummaryResponse(order)
	}
	return responses
}

func ToOrderDetailResponse(orderData *orderapp.Order) dtoresponse.OrderResponse {
	var items []dtoresponse.OrderItem
	for _, item := range orderData.OrderItems {
		items = append(items, dtoresponse.OrderItem{
			ProductName: item.Name,
			Note:        item.Note,
			Quantity:    item.Qty,
			Price:       item.Price.StringFixed(2),
			SubTotal:    item.SubTotal.StringFixed(2),
		})
	}

	return dtoresponse.OrderResponse{
		OrderID:       orderData.Order.OrderNumber,
		OrderNumber:   orderData.Order.QueueNumber,
		TableNumber:   orderData.Order.TableNumber,
		CustomerName:  orderData.Order.CustomerName,
		CustomerEmail: orderData.Order.CustomerEmail,
		CustomerPhone: orderData.Order.CustomerPhone,
		OrderType:     orderData.Order.OrderType,
		GrandTotal:    orderData.Order.Amount.StringFixed(2),
		CreatedAt:     orderData.Order.CreatedAt,
		Items:         items,
	}
}
