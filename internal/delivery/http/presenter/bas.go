package presenter

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/bas"
)

func ToBasResponse(entity *bas.Bas) *dtoresponse.BasResponse {
	return &dtoresponse.BasResponse{
		ID:   entity.ID,
		Code: entity.Code,
		Bas:  entity.Bas,
	}
}
