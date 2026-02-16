package orderstatushistory

import "time"

type OrderStatusHistory struct {
	ID        string
	OrderID   string
	Status    string
	CreatedAt time.Time
}
