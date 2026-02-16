package repository

import (
	"context"
	"goilerplate/internal/domain/orderitem"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type orderItemRepo struct {
	db *gorm.DB
}

func NewOrderItem(db *gorm.DB) orderitem.Repository {
	return &orderItemRepo{db: db}
}

func (r *orderItemRepo) WithTx(ctx context.Context) orderitem.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewOrderItem(tx)
	}
	return r
}

func (r *orderItemRepo) CreateOrderItem(ctx context.Context, item *orderitem.OrderItem) error {
	oi := &model.OrderItem{
		ID:        item.ID,
		OrderID:   item.OrderID,
		ProductID: item.ProductID,
		Name:      item.Name,
		Price:     item.Price,
		Image:     item.Image,
		Quantity:  item.Qty,
		Subtotal:  item.SubTotal,
		Notes:     item.Note,
		CreatedAt: utils.Now(),
	}

	if err := r.db.WithContext(ctx).Create(oi).Error; err != nil {
		return err
	}

	return nil
}

func (r *orderItemRepo) GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]*orderitem.OrderItem, error) {
	var models []model.OrderItem

	if err := r.db.WithContext(ctx).
		Select("id", "order_id", "product_id", "name", "price", "image", "quantity", "subtotal", "notes").
		Where("order_id = ?", orderID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *orderItemRepo) toDomainEntities(models []model.OrderItem) []*orderitem.OrderItem {
	var items []*orderitem.OrderItem
	for i := range models {
		if entity := r.toDomainEntity(&models[i]); entity != nil {
			items = append(items, entity)
		}
	}
	return items
}

func (r *orderItemRepo) toDomainEntity(m *model.OrderItem) *orderitem.OrderItem {
	if m == nil {
		return nil
	}

	return &orderitem.OrderItem{
		ID:        m.ID,
		OrderID:   m.OrderID,
		ProductID: m.ProductID,
		Name:      m.Name,
		Price:     m.Price,
		Image:     m.Image,
		Qty:       m.Quantity,
		SubTotal:  m.Subtotal,
		Note:      m.Notes,
	}
}
