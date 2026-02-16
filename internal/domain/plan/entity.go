package plan

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Plan struct {
	ID             uuid.UUID
	PlanTypeID     uuid.UUID
	DurationInDays int
	Price          decimal.Decimal
	IsActive       bool
}
