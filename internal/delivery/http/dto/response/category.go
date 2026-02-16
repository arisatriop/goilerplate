package dtoresponse

import "github.com/google/uuid"

type CategoryResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"isActive"`
}

type CategoryWithProductsResponse struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Products []ProductBasicInfoResponse `json:"products"`
}

type ProductBasicInfoResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	ImageURL    string `json:"imageUrl"`
	IsAvailable bool   `json:"isAvailable"`
}
