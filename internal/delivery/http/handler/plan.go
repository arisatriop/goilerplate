package handler

import (
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/domain/plantype"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Plan struct {
	Usecase plantype.Usecase
}

func NewPlan(usecase plantype.Usecase) *Plan {
	return &Plan{
		Usecase: usecase,
	}
}

func (h *Plan) List(ctx *fiber.Ctx) error {
	result, err := h.Usecase.GetList(ctx.UserContext())
	if err != nil {
		return response.HandleError(ctx, err)
	}

	planTypeResponse := presenter.ToPlanTypeListResponse(result)

	return response.Success(ctx, planTypeResponse)
}
