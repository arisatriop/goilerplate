package dtoresponse

import "time"

type OrderResponse struct {
	OrderID       string      `json:"orderId"`
	OrderNumber   string      `json:"orderNumber"`
	TableNumber   *string     `json:"tableNumber"`
	CustomerName  *string     `json:"customerName"`
	CustomerEmail *string     `json:"customerEmail"`
	CustomerPhone *string     `json:"customerPhone"`
	OrderType     string      `json:"orderType"`
	GrandTotal    string      `json:"grandTotal"`
	CreatedAt     time.Time   `json:"createdAt"`
	Items         []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductName string  `json:"productName"`
	Note        *string `json:"note"`
	Quantity    int     `json:"quantity"`
	Price       string  `json:"price"`
	SubTotal    string  `json:"subTotal"`
}

type OrderSummaryResponse struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"orderId"`
	OrderNumber   string    `json:"orderNumber"`
	TableNumber   *string   `json:"tableNumber"`
	CustomerName  *string   `json:"customerName"`
	CustomerEmail *string   `json:"customerEmail"`
	CustomerPhone *string   `json:"customerPhone"`
	OrderType     string    `json:"orderType"`
	OrderStatus   string    `json:"orderStatus"`
	Amount        string    `json:"amount"`
	CreatedAt     time.Time `json:"createdAt"`
}
