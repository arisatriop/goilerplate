package plantyperule

import (
	"fmt"

	"github.com/google/uuid"
)

type PlanTypeRule struct {
	ID         uuid.UUID
	PlanTypeID uuid.UUID
	Rule       string
	RuleValue  string
	IsActive   bool
}

func (r *PlanTypeRule) Max() int {
	var max int
	fmt.Sscanf(r.RuleValue, "%d", &max)
	return max
}
