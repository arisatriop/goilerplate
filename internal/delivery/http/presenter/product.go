package presenter

import (
	"goilerplate/internal/application/productapp"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/product"
)

// ToProductResponse converts a single product entity to basic DTO
func ToProductResponse(entity *product.Product) dtoresponse.ProductResponse {
	return dtoresponse.ProductResponse{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		Price:       entity.Price.String(),
		Image:       entity.Images,
		IsAvailable: entity.IsAvailable,
	}
}

// ToProductListResponse converts multiple product entities to basic DTOs
func ToProductListResponse(entities []*product.Product) []dtoresponse.ProductResponse {
	responses := make([]dtoresponse.ProductResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToProductResponse(entity)
	}
	return responses
}

// ToProductDetailsResponse converts product details with categories and images
func ToProductDetailsResponse(entity *productapp.ProductDetails) dtoresponse.ProductDetailsResponse {
	categoryResp := make([]dtoresponse.CategoryOfProduct, len(entity.Categories))
	for i, cat := range entity.Categories {
		categoryResp[i] = dtoresponse.CategoryOfProduct{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	imageResp := make([]dtoresponse.ImageOfProduct, len(entity.Images))
	for i, img := range entity.Images {
		imageResp[i] = dtoresponse.ImageOfProduct{
			ID:          img.ID,
			FileType:    img.FileType,
			FileStorage: img.FileStorage,
			FileName:    img.FileName,
			FilePath:    img.FilePath,
			FileURL:     img.FileURL,
		}
	}

	return dtoresponse.ProductDetailsResponse{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Desc,
		Price:       entity.Price,
		Image:       entity.Image,
		IsAvailable: entity.IsAvailable,
		IsActive:    entity.IsActive,
		Categories:  categoryResp,
		Images:      imageResp,
	}
}

// ToProductWithCategoryResponse converts a single product with categories
func ToProductWithCategoryResponse(entity *productapp.ProductWithCategory) *dtoresponse.ProductWithCategoryResponse {
	categoryResp := make([]dtoresponse.CategoryOfProduct, len(entity.Categories))
	for i, cat := range entity.Categories {
		categoryResp[i] = dtoresponse.CategoryOfProduct{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	return &dtoresponse.ProductWithCategoryResponse{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Desc,
		Price:       entity.Price,
		Image:       entity.Images,
		IsAvailable: entity.IsAvailable,
		IsActive:    entity.IsActive,
		Categories:  categoryResp,
	}
}

// ToProductListWithCategoriesResponse converts multiple products with categories
func ToProductListWithCategoriesResponse(entities []productapp.ProductWithCategory) []*dtoresponse.ProductWithCategoryResponse {
	responses := make([]*dtoresponse.ProductWithCategoryResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToProductWithCategoryResponse(&entity)
	}
	return responses
}
