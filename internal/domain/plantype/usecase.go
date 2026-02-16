package plantype

import "context"

type Usecase interface {
	FilterHighestPlanType(pt []PlanType) *PlanType
	GetList(ctx context.Context) ([]PlanTypeWithPlans, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) FilterHighestPlanType(pt []PlanType) *PlanType {
	// Priority map: higher number = higher priority
	priority := map[string]int{
		SubscriptionEnterpriseCode: 4,
		SubscriptionPlatinumCode:   3,
		SubscriptionProCode:        2,
		SubscriptionBasicCode:      1,
	}

	var highest *PlanType
	highestPriority := 0

	for i := range pt {
		if p, exists := priority[pt[i].Code]; exists && p > highestPriority {
			highestPriority = p
			highest = &pt[i]
		}
	}

	return highest
}

func (uc *usecase) GetList(ctx context.Context) ([]PlanTypeWithPlans, error) {
	return uc.repo.GetListWithPlans(ctx)
}
