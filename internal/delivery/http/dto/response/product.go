package dtoresponse

import "github.com/google/uuid"

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       string    `json:"price"`
	Image       *string   `json:"image"`
	IsAvailable bool      `json:"isAvailable"`
}

type ProductWithCategoryResponse struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Price       string              `json:"price"`
	Image       *string             `json:"image"`
	IsAvailable bool                `json:"isAvailable"`
	IsActive    bool                `json:"isActive"`
	Categories  []CategoryOfProduct `json:"categories"`
}

type ProductDetailsResponse struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Image       *string             `json:"image"`
	Price       string              `json:"price"`
	IsAvailable bool                `json:"isAvailable"`
	IsActive    bool                `json:"isActive"`
	Categories  []CategoryOfProduct `json:"categories"`
	Images      []ImageOfProduct    `json:"images"`
}

type CategoryOfProduct struct {
	ID   uuid.UUID `json:"id"`
	Name string
}

type ImageOfProduct struct {
	ID          uuid.UUID `json:"id"`
	FileType    string    `json:"fileType"`
	FileStorage string    `json:"fileStorage"`
	FileName    string    `json:"fileName"`
	FilePath    string    `json:"filePath"`
	FileURL     string    `json:"fileURL"`
}
