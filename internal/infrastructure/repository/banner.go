package repository

import (
	"context"
	"goilerplate/internal/domain/banner"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bannerRepo struct {
	db *gorm.DB
}

func NewBanner(db *gorm.DB) banner.Repository {
	return &bannerRepo{
		db: db,
	}
}

func (r *bannerRepo) WithTx(ctx context.Context) banner.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewBanner(tx)
	}
	return r
}

func (r *bannerRepo) CountBanners(ctx context.Context, filter *banner.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Banner{}).
		Where("deleted_at IS NULL")

	r.applyBannerFilters(query, filter, false)

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *bannerRepo) GetBannerList(ctx context.Context, filter *banner.Filter) ([]*banner.Banner, error) {
	var models []model.Banner

	query := r.db.WithContext(ctx).
		Select("id", "store_id", "file_type", "file_storage", "file_name", "file_path", "file_url", "is_active").
		Where("deleted_at IS NULL")

	r.applyBannerFilters(query, filter, true)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(models), nil
}

func (r *bannerRepo) GetBannerByID(ctx context.Context, id uuid.UUID, filter *banner.Filter) (*banner.Banner, error) {
	var m model.Banner

	query := r.db.WithContext(ctx).
		Select("id", "store_id", "file_type", "file_storage", "file_name", "file_path", "file_url", "is_active").
		Where("id = ? AND deleted_at IS NULL", id)

	r.applyBannerFilters(query, filter, false)

	if err := query.First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntity(&m), nil
}

func (r *bannerRepo) CreateBanner(ctx context.Context, banner *banner.Banner) error {
	b := &model.Banner{
		ID:          banner.ID,
		StoreID:     banner.StoreID,
		Filetype:    banner.Filetype,
		FileStorage: banner.FileStorage,
		Filename:    banner.Filename,
		Filepath:    banner.Filepath,
		Fileurl:     banner.FileURL,
		IsActive:    banner.IsActive,
	}

	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		return err
	}

	return nil
}

func (r *bannerRepo) DeleteBanner(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Banner{}, id).Error
}

func (r *bannerRepo) UpdateToggleActive(ctx context.Context, banner *banner.Banner, updatedBy string) error {
	// Use map with explicit audit fields to ensure only these fields are updated
	updates := map[string]interface{}{
		"is_active":  banner.IsActive,
		"updated_by": updatedBy,
	}

	return r.db.WithContext(ctx).
		Model(&model.Banner{}).
		Where("id = ?", banner.ID).
		Updates(updates).Error
}

func (r *bannerRepo) applyBannerFilters(query *gorm.DB, filter *banner.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.StoreID != nil {
		query.Where("store_id = ?", *filter.StoreID)
	}

	if filter.IsActive != nil {
		query.Where("is_active = ?", *filter.IsActive)
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *bannerRepo) toDomainEntities(m []model.Banner) []*banner.Banner {
	var banners []*banner.Banner
	for i := range m {
		if entity := r.toDomainEntity(&m[i]); entity != nil {
			banners = append(banners, entity)
		}
	}

	return banners
}

func (r *bannerRepo) toDomainEntity(m *model.Banner) *banner.Banner {
	if m == nil {
		return nil
	}

	return &banner.Banner{
		ID:          m.ID,
		StoreID:     m.StoreID,
		Filetype:    m.Filetype,
		FileStorage: m.FileStorage,
		Filename:    m.Filename,
		Filepath:    m.Filepath,
		FileURL:     m.Fileurl,
		IsActive:    m.IsActive,
	}
}
