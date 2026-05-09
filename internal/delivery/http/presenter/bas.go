package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/bas"
)

func ToBasResponse(entity *bas.Bas) *dtoresponse.BasResponse {
	return &dtoresponse.BasResponse{
		ID:   entity.ID,
		Code: entity.Code,
		Name: entity.Name,
	}
}

func ToBasListResponse(entities []*bas.Bas) []*dtoresponse.BasResponse {
	responses := make([]*dtoresponse.BasResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToBasResponse(entity)
	}
	return responses
}
