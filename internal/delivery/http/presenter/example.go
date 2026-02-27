package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/example"
)

// ToExampleResponse converts a single example entity to DTO
func ToExampleResponse(entity *example.Example) *dtoresponse.ExampleResponse {
	return &dtoresponse.ExampleResponse{
		ID:      entity.ID,
		Code:    entity.Code,
		Example: entity.Example,
	}
}

// ToExampleListResponse converts multiple example entities to DTOs
func ToExampleListResponse(entities []*example.Example) []*dtoresponse.ExampleResponse {
	responses := make([]*dtoresponse.ExampleResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToExampleResponse(entity)
	}
	return responses
}
