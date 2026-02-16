package banner

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	WithTx(ctx context.Context) Repository

	CountBanners(ctx context.Context, filter *Filter) (int64, error)
	GetBannerByID(ctx context.Context, id uuid.UUID, filter *Filter) (*Banner, error)
	GetBannerList(ctx context.Context, filter *Filter) ([]*Banner, error)

	CreateBanner(ctx context.Context, banner *Banner) error
	DeleteBanner(ctx context.Context, id uuid.UUID) error
	UpdateToggleActive(ctx context.Context, banner *Banner, updatedBy string) error
}
