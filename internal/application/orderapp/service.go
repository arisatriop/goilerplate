package orderapp

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/orderitem"
	"goilerplate/internal/domain/orderstatushistory"
	"goilerplate/internal/domain/transaction"
	"goilerplate/pkg/utils"
)

type ApplicationService interface {
	CreateOrder(ctx context.Context, order Order) (*order.Order, error)
	GetOrderDetail(ctx context.Context, orderID string, storeID string) (*Order, error)
}

type applicationService struct {
	txManager          transaction.Transaction
	orderUC            order.Usecase
	orderRepo          order.Repository
	orderItemRepo      orderitem.Repository
	orderStatusHistory orderstatushistory.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	orderUC order.Usecase,
	orderRepo order.Repository,
	orderItemRepo orderitem.Repository,
	orderStatusHistory orderstatushistory.Repository,
) ApplicationService {
	return &applicationService{
		txManager:          txManager,
		orderUC:            orderUC,
		orderRepo:          orderRepo,
		orderItemRepo:      orderItemRepo,
		orderStatusHistory: orderStatusHistory,
	}
}

func (s *applicationService) CreateOrder(ctx context.Context, orderData Order) (*order.Order, error) {
	var err error
	var createdOrder *order.Order

	ordType := order.OrderTypeDineIn
	orderData.Order.OrderNumber = s.orderUC.GenerateOrderNumber(ordType)
	orderData.Order.QueueNumber = s.orderUC.GenerateQueueNumber(ordType)
	orderData.Order.OrderType = ordType
	orderData.Order.OrderStatus = order.OrderStatusProcessing

	if err := s.txManager.Do(ctx, func(ctx context.Context) error {

		txOrder := s.orderRepo.WithTx(ctx)
		txOrderItem := s.orderItemRepo.WithTx(ctx)
		txOrderStatusHistory := s.orderStatusHistory.WithTx(ctx)

		createdOrder, err = txOrder.CreateOrder(ctx, orderData.Order)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range orderData.OrderItems {
			item.OrderID = createdOrder.ID
			err := txOrderItem.CreateOrderItem(ctx, &item)
			if err != nil {
				return fmt.Errorf("failed to create order items: %w", err)
			}
		}

		orderStatusHistory := &orderstatushistory.OrderStatusHistory{
			OrderID:   createdOrder.ID,
			Status:    "Processing",
			CreatedAt: utils.Now(),
		}
		err = txOrderStatusHistory.CreateOrderStatusHistory(ctx, orderStatusHistory)
		if err != nil {
			return fmt.Errorf("failed to create order status history: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return createdOrder, nil
}

func (s *applicationService) GetOrderDetail(ctx context.Context, orderID string, storeID string) (*Order, error) {

	ord, err := s.orderRepo.GetOrderByID(ctx, orderID, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	orderItems, err := s.orderItemRepo.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	result := &Order{
		Order:      ord,
		OrderItems: make([]orderitem.OrderItem, len(orderItems)),
	}

	for i, item := range orderItems {
		result.OrderItems[i] = *item
	}

	return result, nil
}
