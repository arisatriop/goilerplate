package banner

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/banner"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/domain/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/utils"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
)

var (
	boolTrue = true
)

type ApplicationService interface {
	CreateBannerWithFiles(ctx context.Context, storeID uuid.UUID, files []*multipart.FileHeader) error
	UpdateToggleActive(ctx context.Context, banner *banner.Banner) (bool, error)
}

type applicationService struct {
	txManager         transaction.Transaction
	fileSystemManager *filesystem.Manager
	bannerUc          banner.Usecase
	bannerRepo        banner.Repository
	plantypeUc        plantype.Usecase
	plantypeRepo      plantype.Repository
	plantypeRuleUC    plantyperule.Usecase
	plantypeRuleRepo  plantyperule.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	fileSystemManager *filesystem.Manager,
	bannerUc banner.Usecase,
	bannerRepo banner.Repository,
	plantypeUc plantype.Usecase,
	plantypeRepo plantype.Repository,
	plantypeRuleUC plantyperule.Usecase,
	plantypeRuleRepo plantyperule.Repository,
) ApplicationService {
	return &applicationService{
		txManager:         txManager,
		fileSystemManager: fileSystemManager,
		bannerUc:          bannerUc,
		bannerRepo:        bannerRepo,
		plantypeUc:        plantypeUc,
		plantypeRepo:      plantypeRepo,
		plantypeRuleUC:    plantypeRuleUC,
		plantypeRuleRepo:  plantypeRuleRepo,
	}
}

func (s *applicationService) CreateBannerWithFiles(ctx context.Context, storeID uuid.UUID, files []*multipart.FileHeader) error {
	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUc.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	existingActiveBanner, err := s.bannerRepo.GetBannerList(ctx, &banner.Filter{StoreID: &storeID, IsActive: &boolTrue})
	if err != nil {
		return fmt.Errorf("failed to get existing banners: %v", err)
	}

	// Use transaction for creating banners
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txBannerRepo := s.bannerRepo.WithTx(txCtx)

		uploadOptions := filesystem.UploadOptions{
			Path:             fmt.Sprintf("stores/%s/banners/", storeID.String()),
			Public:           true,
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/jpg"},
		}

		currentActiveBanners := len(existingActiveBanner)
		maxBanners := s.plantypeRuleUC.GetMaxBannerFromRules(rules)

		for _, file := range files {
			uploadResult, err := s.fileSystemManager.Upload(file, uploadOptions)
			if err != nil {
				return fmt.Errorf("failed to upload banner file %s: %w", file.Filename, err)
			}

			isActiveBanner := maxBanners == 0 || currentActiveBanners < maxBanners
			if isActiveBanner {
				currentActiveBanners++ // Increment count for next iteration
			}

			bannerEntity := &banner.Banner{
				ID:          uuid.New(),
				StoreID:     storeID,
				Filetype:    uploadResult.MimeType,
				FileStorage: string(uploadResult.Driver),
				Filename:    uploadResult.Filename,
				Filepath:    uploadResult.Path,
				FileURL:     &uploadResult.URL,
				IsActive:    isActiveBanner, // Set based on limit check
			}

			err = txBannerRepo.CreateBanner(txCtx, bannerEntity)
			if err != nil {
				return fmt.Errorf("failed to create banner: %w", err)
			}
		}

		return nil
	})
}

func (s *applicationService) UpdateToggleActive(ctx context.Context, b *banner.Banner) (bool, error) {
	bnrs, err := s.bannerRepo.GetBannerByID(ctx, b.ID, &banner.Filter{StoreID: &b.StoreID})
	if err != nil {
		return false, fmt.Errorf("failed to get banner by id: %w", err)
	}
	if bnrs == nil {
		return false, utils.Error(http.StatusNotFound, "banner not found")
	}

	bnrs.IsActive = !bnrs.IsActive
	if !bnrs.IsActive {
		if err := s.bannerRepo.UpdateToggleActive(ctx, bnrs, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
			return false, fmt.Errorf("failed to update banner: %w", err)
		}
		return bnrs.IsActive, nil
	}

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, bnrs.StoreID)
	if err != nil {
		return false, fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUc.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return false, fmt.Errorf("failed to get plan type rules: %v", err)
	}

	currentActiveBanners, err := s.bannerRepo.GetBannerList(ctx, &banner.Filter{StoreID: &b.StoreID, IsActive: &boolTrue})
	if err != nil {
		return false, fmt.Errorf("failed to get current active banners: %w", err)
	}

	maxBanners := s.plantypeRuleUC.GetMaxBannerFromRules(rules)
	if maxBanners > 0 && len(currentActiveBanners) >= maxBanners {
		return false, utils.Error(http.StatusBadRequest, banner.MsgMaxActiveBannersReached)
	}

	if err := s.bannerRepo.UpdateToggleActive(ctx, bnrs, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
		return false, fmt.Errorf("failed to update banner: %w", err)
	}

	return bnrs.IsActive, nil
}
