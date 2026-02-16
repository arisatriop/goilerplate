package dtorequest

type OrderCreateRequest struct {
	CustomerName *string                  `json:"customeName"`
	CutomerEmail *string                  `json:"customerEmail"`
	CutomerPhone *string                  `json:"customerPhone"`
	TableNumber  *string                  `json:"tableNumber"`
	Notes        *string                  `json:"notes"`
	GrandTotal   string                   `json:"grandTotal"`
	Items        []OrderItemCreateRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderItemCreateRequest struct {
	ProductID   string  `json:"productId" validate:"required"`
	ProductName string  `json:"productName" validate:"required"`
	Image       *string `json:"image"`
	Quantity    int     `json:"qty" validate:"required,min=1"`
	Price       string  `json:"price" validate:"required"`
	Notes       *string `json:"notes"`
	SubTotal    string  `json:"subTotal"`
}
