package category

import "github.com/google/uuid"

type Category struct {
	ID       uuid.UUID
	Name     string
	IsActive bool
	StoreID  uuid.UUID
}

type CategoryWithProducts struct {
	ID                 string
	Name               string
	ProductID          string
	ProductName        string
	ProductDesc        string
	ProductPrice       string
	ProductImages      string
	ProductIsAvailable bool
}

type CategoryWithProductsGrouping struct {
	ID       string
	Name     string
	Products []ProductDetails
}

type ProductDetails struct {
	ID          string
	Name        string
	Desc        string
	Price       string
	ImageURL    string
	IsAvailable bool
}
