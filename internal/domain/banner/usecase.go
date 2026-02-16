package banner

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/utils"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
)

type Usecase interface {
	GetList(ctx context.Context, filter *Filter) ([]*Banner, int64, error)
	Delete(ctx context.Context, id uuid.UUID, filter *Filter) error
	Upload(ctx context.Context, form []*multipart.FileHeader) ([]Banner, error)
}

type usecase struct {
	repo          Repository
	filesystemMgr *filesystem.Manager
}

func NewUseCase(repo Repository, filesystemMgr *filesystem.Manager) Usecase {
	return &usecase{
		repo:          repo,
		filesystemMgr: filesystemMgr,
	}
}

func (uc *usecase) GetList(ctx context.Context, filter *Filter) ([]*Banner, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	banners, err := uc.repo.GetBannerList(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get banners: %v", err)
	}

	total, err := uc.repo.CountBanners(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count banners: %v", err)
	}

	return banners, total, nil
}

func (uc *usecase) Delete(ctx context.Context, id uuid.UUID, filter *Filter) error {

	banner, err := uc.repo.GetBannerByID(ctx, id, filter)
	if err != nil {
		return fmt.Errorf("failed to get banner by id: %v", err)
	}
	if banner == nil {
		return utils.Error(http.StatusBadRequest, "banner not found")
	}

	if err := uc.repo.DeleteBanner(ctx, id); err != nil {
		return fmt.Errorf("failed to delete banner: %v", err)
	}
	return nil
}

func (uc *usecase) Upload(ctx context.Context, files []*multipart.FileHeader) ([]Banner, error) {
	storeIDStr, _ := ctx.Value(constants.ContextKeyStoreID).(string)

	if len(files) == 0 {
		return nil, utils.Error(http.StatusBadRequest, MsgMinOneFileRequired)
	}

	var banners []Banner
	for _, file := range files {
		opts := filesystem.UploadOptions{
			Path:     fmt.Sprintf("stores/%s/banners", storeIDStr),
			Filename: file.Filename,
			Public:   true,
		}
		result, err := uc.filesystemMgr.Upload(file, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %v", err)
		}

		banners = append(banners, Banner{
			Filetype:    result.MimeType,
			FileStorage: string(result.Driver),
			Filepath:    result.Path,
			Filename:    result.Filename,
			FileURL:     &result.URL,
		})
	}

	return banners, nil
}
