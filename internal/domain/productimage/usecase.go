package productimage

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/utils"
	"mime/multipart"
	"net/http"
)

type Usecase interface {
	Upload(ctx context.Context, form []*multipart.FileHeader) ([]ProductImage, error)
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

func (uc *usecase) Upload(ctx context.Context, files []*multipart.FileHeader) ([]ProductImage, error) {
	storeIDStr, _ := ctx.Value(constants.ContextKeyStoreID).(string)

	if len(files) == 0 {
		return nil, utils.Error(http.StatusBadRequest, MsgMinOneFileRequired)
	}

	var productImages []ProductImage
	for _, file := range files {
		opts := filesystem.UploadOptions{
			Path:     fmt.Sprintf("stores/%s/products", storeIDStr),
			Filename: file.Filename,
			Public:   true,
		}
		result, err := uc.filesystemMgr.Upload(file, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %v", err)
		}

		productImages = append(productImages, ProductImage{
			FileType:    result.MimeType,
			FileStorage: string(result.Driver),
			FilePath:    result.Path,
			FileName:    result.Filename,
			FileURL:     result.URL,
		})
	}

	return productImages, nil
}
