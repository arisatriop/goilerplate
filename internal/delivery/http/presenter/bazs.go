package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/bazs"
)

func ToBazsResponse(entity *bazs.Bazs) *dtoresponse.BazsResponse {
	return &dtoresponse.BazsResponse{
		ID:   entity.ID,
		Code: entity.Code,
		Name: entity.Name,
	}
}

func ToBazsListResponse(entities []*bazs.Bazs) []*dtoresponse.BazsResponse {
	responses := make([]*dtoresponse.BazsResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToBazsResponse(entity)
	}
	return responses
}
