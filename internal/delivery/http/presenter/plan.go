package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/plantype"
)

func ToPlanTypeListResponse(entities []plantype.PlanTypeWithPlans) []dtoresponse.PlanTypeResponse {
	responses := make([]dtoresponse.PlanTypeResponse, len(entities))
	for i, entity := range entities {
		responses[i] = dtoresponse.PlanTypeResponse{
			ID:       entity.ID,
			Code:     entity.Code,
			Name:     entity.Name,
			IsActive: entity.IsActive,
			Plans:    toPlanItemListResponse(entity.Plans),
		}
	}
	return responses
}

func toPlanItemListResponse(items []plantype.PlanItem) []dtoresponse.PlanItemResponse {
	responses := make([]dtoresponse.PlanItemResponse, len(items))
	for i, item := range items {
		responses[i] = dtoresponse.PlanItemResponse{
			ID:             item.ID,
			DurationInDays: item.DurationInDays,
			Price:          item.Price.StringFixed(2),
			IsActive:       item.IsActive,
		}
	}
	return responses
}
