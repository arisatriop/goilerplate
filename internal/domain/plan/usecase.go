package plan

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"strings"
)

type Usecase interface {
	GetBasicPlan(ctx context.Context) (*Plan, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) GetBasicPlan(ctx context.Context) (*Plan, error) {
	// The basic plan rules not defined yet
	// Assuming "P001B" is the code for the basic plan type
	plans, err := uc.repo.GetPlanByPlanTypeCode(ctx, BasicPlanTypeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %v", err)
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf("failed to get plan: %v", strings.ToLower(constants.MsgFeatureNotImplemented))
	}

	return &plans[0], nil
}
