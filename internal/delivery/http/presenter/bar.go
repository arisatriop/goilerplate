package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/bar"
)

// ToBarResponse converts a single bar entity to DTO
func ToBarResponse(entity *bar.Bar) *dtoresponse.BarResponse {
	return &dtoresponse.BarResponse{
		ID:      entity.ID,
		Code:    entity.Code,
		Bar: entity.Bar,
	}
}

// ToBarListResponse converts multiple bar entities to DTOs
func ToBarListResponse(entities []*bar.Bar) []*dtoresponse.BarResponse {
	responses := make([]*dtoresponse.BarResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToBarResponse(entity)
	}
	return responses
}
