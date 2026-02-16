package plantype

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PlanType struct {
	ID       uuid.UUID
	Code     string
	Name     string
	IsActive bool
}

type PlanTypeWithPlans struct {
	ID       uuid.UUID
	Code     string
	Name     string
	IsActive bool
	Plans    []PlanItem
}

type PlanItem struct {
	ID             uuid.UUID
	DurationInDays int
	Price          decimal.Decimal
	IsActive       bool
}
