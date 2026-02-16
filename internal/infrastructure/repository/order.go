package repository

import (
	"context"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderRepo struct {
	db *gorm.DB
}

func NewOrder(db *gorm.DB) order.Repository {
	return &orderRepo{db: db}
}

func (r *orderRepo) WithTx(ctx context.Context) order.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewOrder(tx)
	}
	return r
}

func (r *orderRepo) CreateOrder(ctx context.Context, ord *order.Order) (*order.Order, error) {

	o := &model.Order{
		ID:            uuid.New().String(),
		OrderNumber:   ord.OrderNumber,
		QueueNumber:   &ord.QueueNumber,
		TableNumber:   ord.TableNumber,
		OrderType:     ord.OrderType,
		OrderStatus:   ord.OrderStatus,
		Notes:         ord.Notes,
		Amount:        ord.Amount,
		StoreID:       ord.StoreID,
		CustomerName:  ord.CustomerName,
		CustomerPhone: ord.CustomerPhone,
		CustomerEmail: ord.CustomerEmail,
		CreatedAt:     utils.Now(),
		UpdatedAt:     utils.Now(),
	}

	if err := r.db.WithContext(ctx).Create(o).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntity(o), nil
}

func (r *orderRepo) GetOrderByID(ctx context.Context, orderID string, storeID string) (*order.Order, error) {
	var m model.Order

	if err := r.db.WithContext(ctx).
		Select("id", "order_number", "queue_number", "table_number", "order_type", "order_status", "notes", "amount", "store_id", "customer_name", "customer_phone", "customer_email", "created_at").
		Where("id = ? AND store_id = ?", orderID, storeID).
		First(&m).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntity(&m), nil
}

func (r *orderRepo) GetListOrder(ctx context.Context, filter *order.Filter) ([]*order.Order, error) {
	var models []model.Order

	query := r.db.WithContext(ctx).
		Select("id", "order_number", "queue_number", "table_number", "order_type", "order_status", "notes", "amount", "store_id", "customer_name", "customer_phone", "customer_email", "created_at").
		Order("created_at DESC")
	r.applyOrderFilters(query, filter, true)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *orderRepo) CountOrders(ctx context.Context, filter *order.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.Order{})
	r.applyOrderFilters(query, filter, false)

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *orderRepo) applyOrderFilters(query *gorm.DB, filter *order.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	// StoreID is required - always filter by store
	if filter.StoreID != "" {
		query.Where("store_id = ?", filter.StoreID)
	}

	// Search keyword in queue number OR customer name
	if filter.Keyword != nil && *filter.Keyword != "" {
		keyword := "%" + *filter.Keyword + "%"
		query.Where("LOWER(queue_number) LIKE LOWER(?) OR LOWER(customer_name) LIKE LOWER(?)", keyword, keyword)
	}

	// Apply pagination
	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *orderRepo) toDomainEntities(models []model.Order) []*order.Order {
	var orders []*order.Order
	for i := range models {
		if entity := r.toDomainEntity(&models[i]); entity != nil {
			orders = append(orders, entity)
		}
	}
	return orders
}

func (r *orderRepo) toDomainEntity(m *model.Order) *order.Order {
	if m == nil {
		return nil
	}

	return &order.Order{
		ID:            m.ID,
		OrderNumber:   m.OrderNumber,
		QueueNumber:   *m.QueueNumber,
		TableNumber:   m.TableNumber,
		OrderType:     m.OrderType,
		OrderStatus:   m.OrderStatus,
		Notes:         m.Notes,
		Amount:        m.Amount,
		StoreID:       m.StoreID,
		CustomerName:  m.CustomerName,
		CustomerPhone: m.CustomerPhone,
		CustomerEmail: m.CustomerEmail,
		CreatedAt:     m.CreatedAt,
	}
}
