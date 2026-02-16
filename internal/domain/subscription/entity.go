package subscription

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Subscription struct {
	ID       uuid.UUID
	StoreID  uuid.UUID
	PlanID   uuid.UUID
	StarDate *time.Time
	EndDate  *time.Time
	Price    decimal.Decimal
	Status   string
	IsActive bool
}
