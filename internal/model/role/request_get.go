package role

import "goilerplate/internal/model"

type GetRequest struct {
	model.Params
}

func GetParams() *GetRequest {
	return &GetRequest{
		Params: model.DefaultParams(),
	}
}
