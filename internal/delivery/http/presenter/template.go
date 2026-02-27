package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/template"
)

// ToTemplateResponse converts a single template entity to DTO
func ToTemplateResponse(entity *template.Template) *dtoresponse.TemplateResponse {
	return &dtoresponse.TemplateResponse{
		ID:       entity.ID,
		Code:     entity.Code,
		Template: entity.Template,
	}
}

// ToExampleListResponse converts multiple example entities to DTOs
func ToTemplateListResponse(entities []*template.Template) []*dtoresponse.TemplateResponse {
	responses := make([]*dtoresponse.TemplateResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToTemplateResponse(entity)
	}
	return responses
}
