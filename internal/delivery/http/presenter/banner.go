package presenter

import (
	"goilerplate/internal/domain/banner"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
)

// ToBannerResponse converts a single banner entity to DTO
func ToBannerResponse(entity *banner.Banner) *dtoresponse.BannerResponse {
	return &dtoresponse.BannerResponse{
		ID:          entity.ID,
		Filetype:    entity.Filetype,
		FileStorage: entity.FileStorage,
		Filename:    entity.Filename,
		Filepath:    entity.Filepath,
		FileURL:     entity.FileURL,
		IsActive:    entity.IsActive,
	}
}

// ToBannerListResponse converts multiple banner entities to DTOs
func ToBannerListResponse(entities []*banner.Banner) []*dtoresponse.BannerResponse {
	responses := make([]*dtoresponse.BannerResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToBannerResponse(entity)
	}
	return responses
}

// ToBannerUploadResponse converts a single banner entity to upload response DTO
func ToBannerUploadResponse(entity *banner.Banner) dtoresponse.BannerUploadResponse {
	return dtoresponse.BannerUploadResponse{
		Filetype:    entity.Filetype,
		FileStorage: entity.FileStorage,
		Filename:    entity.Filename,
		Filepath:    entity.Filepath,
		FileURL:     entity.FileURL,
	}
}

// ToBannerUploadListResponse converts multiple banner entities to upload response DTOs
func ToBannerUploadListResponse(entities []banner.Banner) []dtoresponse.BannerUploadResponse {
	responses := make([]dtoresponse.BannerUploadResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToBannerUploadResponse(&entity)
	}
	return responses
}
