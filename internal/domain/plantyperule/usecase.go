package plantyperule

type Usecase interface {
	GetMaxBannerFromRules(rules []PlanTypeRule) int
	GetMaxCategoryFromRules(rules []PlanTypeRule) int
	GetMaxProductFromRules(rules []PlanTypeRule) int
	GetMaxProductPerCategoryFromRules(rules []PlanTypeRule) int
	GetMaxImagesFromRules(rules []PlanTypeRule) int
	GetMaxImagePerProductFromRules(rules []PlanTypeRule) int
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (uc *usecase) GetMaxBannerFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitBanner {
			return rule.Max()
		}
	}
	return 0
}

func (uc *usecase) GetMaxCategoryFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitCategory {
			return rule.Max()
		}
	}
	return 0
}

func (uc *usecase) GetMaxProductFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitProduct {
			return rule.Max()
		}
	}
	return 0
}

func (uc *usecase) GetMaxProductPerCategoryFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitProductPerCategory {
			return rule.Max()
		}
	}
	return 0
}

func (uc *usecase) GetMaxImagesFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitImage {
			return rule.Max()
		}
	}
	return 0
}

func (uc *usecase) GetMaxImagePerProductFromRules(rules []PlanTypeRule) int {
	for _, rule := range rules {
		if rule.Rule == RuleNameLimitImagePerProduct {
			return rule.Max()
		}
	}
	return 0
}
