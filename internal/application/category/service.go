package category

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/category"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/infrastructure/cache"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

var (
	boolTrue = true
)

type ApplicationService interface {
	CreateCategory(ctx context.Context, c []*category.Category) error
	UpdateToggleActive(ctx context.Context, category *category.Category) (bool, error)
}

type applicationService struct {
	txManager        transaction.Transaction
	cache            *cache.RedisService
	categoryRepo     category.Repository
	plantypeUC       plantype.Usecase
	plantypeRepo     plantype.Repository
	plantypeRuleUC   plantyperule.Usecase
	plantypeRuleRepo plantyperule.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	cache *cache.RedisService,
	categoryRepo category.Repository,
	plantypeUC plantype.Usecase,
	plantypeRepo plantype.Repository,
	plantypeRuleUC plantyperule.Usecase,
	plantypeRuleRepo plantyperule.Repository,
) ApplicationService {
	return &applicationService{
		txManager:        txManager,
		cache:            cache,
		categoryRepo:     categoryRepo,
		plantypeUC:       plantypeUC,
		plantypeRepo:     plantypeRepo,
		plantypeRuleUC:   plantypeRuleUC,
		plantypeRuleRepo: plantypeRuleRepo,
	}
}

func (s *applicationService) CreateCategory(ctx context.Context, c []*category.Category) error {

	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	parseUUID, _ := uuid.Parse(storeIDStr)

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, parseUUID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	existingActiveCategory, err := s.categoryRepo.GetCategoryList(ctx, &category.Filter{StoreID: &parseUUID, IsActive: &boolTrue})
	if err != nil {
		return fmt.Errorf("failed to get category list: %v", err)
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txCategoryRepo := s.categoryRepo.WithTx(txCtx)

		currentActiveCategories := len(existingActiveCategory)
		maxCategories := s.plantypeRuleUC.GetMaxCategoryFromRules(rules)

		isActive := maxCategories == 0 || currentActiveCategories < maxCategories
		if isActive {
			currentActiveCategories++
		}

		for _, v := range c {
			categoryEntity := &category.Category{
				Name:     v.Name,
				StoreID:  parseUUID,
				IsActive: isActive,
			}

			err := txCategoryRepo.CreateCategory(txCtx, categoryEntity)
			if err != nil {
				return fmt.Errorf("failed to create category: %v", err)
			}
		}

		return nil
	})
}

func (s *applicationService) UpdateToggleActive(ctx context.Context, e *category.Category) (bool, error) {
	cat, err := s.categoryRepo.GetCategoryByID(ctx, e.ID, &category.Filter{StoreID: &e.StoreID})
	if err != nil {
		return false, fmt.Errorf("failed to get category by ID: %v", err)
	}

	if cat == nil {
		return false, utils.Error(http.StatusNotFound, "category not found")
	}

	cat.IsActive = !cat.IsActive
	if !cat.IsActive {
		if err := s.categoryRepo.UpdateToggleActive(ctx, cat, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
			return false, fmt.Errorf("failed to update category: %v", err)
		}
		return cat.IsActive, nil
	}

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, cat.StoreID)
	if err != nil {
		return false, fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return false, fmt.Errorf("failed to get plan type rules: %v", err)
	}

	currentActiveCategories, err := s.categoryRepo.GetCategoryList(ctx, &category.Filter{StoreID: &cat.StoreID, IsActive: &boolTrue})
	if err != nil {
		return false, fmt.Errorf("failed to get active category list: %v", err)
	}

	maxCategories := s.plantypeRuleUC.GetMaxCategoryFromRules(rules)
	if maxCategories > 0 && len(currentActiveCategories) >= maxCategories {
		return false, utils.Error(http.StatusBadRequest, category.MsgMaxActiveCategoriesReached)
	}

	if err := s.categoryRepo.UpdateToggleActive(ctx, cat, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
		return false, fmt.Errorf("failed to update banner: %w", err)
	}

	return cat.IsActive, nil
}
