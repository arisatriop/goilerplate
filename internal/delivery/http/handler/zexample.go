package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/zexample"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Example struct {
	Validator *validator.Validate
	Usecase   zexample.Usecase
}

func NewExample(validator *validator.Validate, usecase zexample.Usecase) *Example {
	return &Example{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *Example) Create(ctx *fiber.Ctx) error {
	var req dtorequest.ExampleCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &zexample.Zexample{
		Code:    req.Code,
		Example: req.Example,
	}

	err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil, response.WithMessage(zexample.MsgExampleCreatedSuccessfully))
}

func (h *Example) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var req dtorequest.ExampleUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &zexample.Zexample{
		ID:      id,
		Code:    req.Code,
		Example: req.Example,
	}

	err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(zexample.MsgExampleUpdatedSuccessfully))
}

func (h *Example) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity := &zexample.Zexample{
		ID: id,
	}

	err := h.Usecase.Delete(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Example) List(ctx *fiber.Ctx) error {
	var req dtorequest.ExampleListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	filter := request.ToExampleFilter(&req, ctx)

	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	exampleResponses := presenter.ToExampleListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(exampleResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse, response.WithMessage(zexample.MsgExampleListFetchSuccessfully))
}

func (h *Example) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	exampleResponse := presenter.ToExampleResponse(entity)

	return response.Success(ctx, exampleResponse, response.WithMessage(zexample.MsgExampleFetchedSuccessfully))
}
