package model

import (
	"time"

	"github.com/google/uuid"
)

type PlanTypeRule struct {
	ID         uuid.UUID
	PlanTypeID uuid.UUID
	Rule       string
	RuleValue  string
	IsActive   bool
	CreatedBy  string
	UpdatedBy  string
	DeletedBy  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func (PlanTypeRule) TableName() string {
	return "plan_type_rules"
}
