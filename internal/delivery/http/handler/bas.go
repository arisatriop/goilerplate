package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/domain/bas"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Bas struct {
	Validator *validator.Validate
	Usecase   bas.Usecase
}

func NewBas(validator *validator.Validate, usecase bas.Usecase) *Bas {
	return &Bas{
		Validator: validator,
		Usecase:   usecase,
	}
}

// @Summary      Submit bas
// @Tags         bass
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.BasCreateRequest  true  "Bas data"
// @Success      201      {object}  response.BaseResponse{data=dtoresponse.BasResponse}
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      409      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bass [post]
func (h *Bas) Create(ctx *fiber.Ctx) error {
	var req dtorequest.BasCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bas.Bas{
		Code: req.Code,
		Bas:  req.Bas,
	}

	created, err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, presenter.ToBasResponse(created), response.WithMessage(bas.MsgBasCreatedSuccessfully))
}
