package dtoresponse

import "github.com/google/uuid"

type BannerResponse struct {
	ID          uuid.UUID `json:"id"`
	Filetype    string    `json:"fileType"`
	FileStorage string    `json:"fileStorage"`
	Filename    string    `json:"fileName"`
	Filepath    string    `json:"filePath"`
	FileURL     *string   `json:"fileUrl"`
	IsActive    bool      `json:"isActive"`
}

type BannerUploadResponse struct {
	Filetype    string  `json:"fileType"`
	FileStorage string  `json:"fileStorage"`
	Filename    string  `json:"fileName"`
	Filepath    string  `json:"filePath"`
	FileURL     *string `json:"fileUrl"`
}
