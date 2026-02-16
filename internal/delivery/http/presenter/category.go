package presenter

import (
	"goilerplate/internal/domain/category"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
)

// ToCategoryResponse converts a single category entity to DTO
func ToCategoryResponse(entity *category.Category) *dtoresponse.CategoryResponse {
	return &dtoresponse.CategoryResponse{
		ID:       entity.ID,
		Name:     entity.Name,
		IsActive: entity.IsActive,
	}
}

// ToCategoryListResponse converts multiple category entities to DTOs
func ToCategoryListResponse(entities []*category.Category) []*dtoresponse.CategoryResponse {
	responses := make([]*dtoresponse.CategoryResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToCategoryResponse(entity)
	}
	return responses
}

// ToCategoryWithProductsResponse transforms flat category-product data into nested structure
// This handles the complex grouping logic that was previously in the handler
func ToCategoryWithProductsResponse(data []category.CategoryWithProducts) []dtoresponse.CategoryWithProductsResponse {
	// Group products by category while preserving order
	categoryMap := make(map[string]*dtoresponse.CategoryWithProductsResponse)
	categoryOrder := make([]string, 0)

	for _, item := range data {
		// If category doesn't exist in map, create it
		if _, exists := categoryMap[item.ID]; !exists {
			categoryMap[item.ID] = &dtoresponse.CategoryWithProductsResponse{
				ID:       item.ID,
				Name:     item.Name,
				Products: []dtoresponse.ProductBasicInfoResponse{},
			}
			// Track the order in which categories appear
			categoryOrder = append(categoryOrder, item.ID)
		}

		// Add product to category (only if ProductID is not empty)
		if item.ProductID != "" {
			categoryMap[item.ID].Products = append(categoryMap[item.ID].Products, dtoresponse.ProductBasicInfoResponse{
				ID:          item.ProductID,
				Name:        item.ProductName,
				Description: item.ProductDesc,
				Price:       item.ProductPrice,
				ImageURL:    item.ProductImages,
				IsAvailable: item.ProductIsAvailable,
			})
		}
	}

	// Convert map to slice using the preserved order
	responses := make([]dtoresponse.CategoryWithProductsResponse, 0, len(categoryMap))
	for _, catID := range categoryOrder {
		responses = append(responses, *categoryMap[catID])
	}

	return responses
}
