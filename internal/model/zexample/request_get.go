package zexample

import (
	"golang-clean-architecture/internal/model"
)

type GetRequest struct {
	model.Params
	FieldID string `query:"field_id"`
}

func GetParams() *GetRequest {
	return &GetRequest{
		Params: model.Params{
			Limit:  10,
			Offset: 0,
		},
	}
}
