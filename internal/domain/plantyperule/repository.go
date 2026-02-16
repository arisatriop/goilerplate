package plantyperule

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	GetPlanTypeRuleByPlanTypeID(ctx context.Context, planTypeID string) ([]PlanTypeRule, error)
}
