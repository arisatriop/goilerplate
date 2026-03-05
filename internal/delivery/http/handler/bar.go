package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/bar"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Bar struct {
	Validator *validator.Validate
	Usecase   bar.Usecase
}

func NewBar(validator *validator.Validate, usecase bar.Usecase) *Bar {
	return &Bar{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *Bar) Create(ctx *fiber.Ctx) error {
	var req dtorequest.BarCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bar.Bar{
		Code:    req.Code,
		Bar: req.Bar,
	}

	err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil, response.WithMessage(bar.MsgBarCreatedSuccessfully))
}

func (h *Bar) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var req dtorequest.BarUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bar.Bar{
		ID:      id,
		Code:    req.Code,
		Bar: req.Bar,
	}

	err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(bar.MsgBarUpdatedSuccessfully))
}

func (h *Bar) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity := &bar.Bar{
		ID: id,
	}

	err := h.Usecase.Delete(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Bar) List(ctx *fiber.Ctx) error {
	var req dtorequest.BarListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	filter := request.ToBarFilter(&req, ctx)

	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	barResponses := presenter.ToBarListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(barResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse, response.WithMessage(bar.MsgBarListFetchSuccessfully))
}

func (h *Bar) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	barResponse := presenter.ToBarResponse(entity)

	return response.Success(ctx, barResponse, response.WithMessage(bar.MsgBarFetchedSuccessfully))
}
