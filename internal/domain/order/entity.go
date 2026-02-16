package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID            string
	OrderNumber   string
	QueueNumber   string
	TableNumber   *string
	OrderType     string
	OrderStatus   string
	Notes         *string
	Amount        decimal.Decimal
	StoreID       string
	CustomerName  *string
	CustomerEmail *string
	CustomerPhone *string
	CreatedAt     time.Time
}
