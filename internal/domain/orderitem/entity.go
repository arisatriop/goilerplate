package orderitem

import "github.com/shopspring/decimal"

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Name      string
	Price     decimal.Decimal
	Image     *string
	Qty       int
	SubTotal  decimal.Decimal
	Note      *string
}
