package plan

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	GetPlanByPlanTypeCode(ctx context.Context, code string) ([]Plan, error)
}
