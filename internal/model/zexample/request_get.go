package zexample

import (
	"goilerplate/internal/model"
)

type GetRequest struct {
	model.Params
	FieldID string `query:"field_id"`
}

func GetParams() *GetRequest {
	return &GetRequest{
		Params: model.DefaultParams(),
	}
}
