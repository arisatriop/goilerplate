package product

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Price       decimal.Decimal
	Images      *string
	StoreID     uuid.UUID
	IsActive    bool
	IsAvailable bool
}

type ProductWithCategories struct {
	Product
	Categories []Category
}

type Category struct {
	ID   uuid.UUID
	Name string
}
