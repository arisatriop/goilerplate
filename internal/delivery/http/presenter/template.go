package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/template"
)

// ToTemplateResponse converts a single template entity to DTO
func ToTemplateResponse(entity *template.Template) *dtoresponse.TemplateResponse {
	panic("Implement me")
}

// ToTemplateListResponse converts multiple example entities to DTOs
func ToTemplateListResponse(entities []*template.Template) []*dtoresponse.TemplateResponse {
	panic("Implement me")
}
