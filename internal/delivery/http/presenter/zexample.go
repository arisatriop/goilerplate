package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/zexample"
)

// ToExampleResponse converts a single example entity to DTO
func ToExampleResponse(entity *zexample.Zexample) *dtoresponse.ExampleResponse {
	return &dtoresponse.ExampleResponse{
		ID:      entity.ID,
		Code:    entity.Code,
		Example: entity.Example,
	}
}

// ToExampleListResponse converts multiple example entities to DTOs
func ToExampleListResponse(entities []*zexample.Zexample) []*dtoresponse.ExampleResponse {
	responses := make([]*dtoresponse.ExampleResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToExampleResponse(entity)
	}
	return responses
}
