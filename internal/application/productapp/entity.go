package productapp

import (
	"goilerplate/internal/domain/category"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/domain/productimage"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	Name       string
	Desc       *string
	Price      decimal.Decimal
	Categories []string
	Images     []productimage.ProductImage
}

type ProductDetails struct {
	ID          uuid.UUID
	Name        string
	Desc        *string
	Price       string
	Image       *string
	IsAvailable bool
	IsActive    bool
	Categories  []category.Category
	Images      []productimage.ProductImage
}

type ProductWithCategory struct {
	ID          uuid.UUID
	Name        string
	Desc        *string
	Price       string
	Images      *string
	IsAvailable bool
	IsActive    bool
	Categories  []category.Category
}

type CategoryWithProducts struct {
	ID       uuid.UUID
	Name     string
	Products []product.Product
}

type SubscriptionRule struct {
	LimitCategory           string
	LimitProduct            string
	LimitProductPerCategory string
	LimitImage              string
	LimitImagePerProduct    string
	LimitBanner             string
	DirectChatByWa          string
	OrderService            string
	PaymentService          string
	TableService            string
}
