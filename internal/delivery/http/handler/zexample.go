package handler

import (
	"errors"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/zexample"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"
	"goilerplate/pkg/utils"

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

	entity := &zexample.Example{
		Code:    req.Code,
		Example: req.Example,
	}

	err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		var clientError *utils.ClientError
		if errors.As(err, &clientError) {
			return response.BadRequest(ctx, clientError.Message, nil)
		}

		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	return response.Created(ctx, nil, response.WithMessage(constants.MsgSuccess))
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

	entity := &zexample.Example{
		ID:      id,
		Code:    req.Code,
		Example: req.Example,
	}

	err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		var clientError *utils.ClientError
		if errors.As(err, &clientError) {
			return response.BadRequest(ctx, clientError.Message, nil)
		}

		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	return response.Success(ctx, nil, response.WithMessage(constants.MsgSuccess))
}

func (h *Example) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity := &zexample.Example{
		ID: id,
	}

	err := h.Usecase.Delete(ctx.UserContext(), entity)
	if err != nil {
		var clientError *utils.ClientError
		if errors.As(err, &clientError) {
			return response.BadRequest(ctx, clientError.Message, nil)
		}

		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	return response.NoContent(ctx)
}

func (h *Example) List(ctx *fiber.Ctx) error {
	var req dtorequest.ExampleListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	filter := zexample.ParseFromRequest(&req, ctx)

	result, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		var clientError *utils.ClientError
		if errors.As(err, &clientError) {
			return response.BadRequest(ctx, clientError.Message, nil)
		}

		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	total, err := h.Usecase.Count(ctx.UserContext(), filter)
	if err != nil {
		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	exampleResponses := make([]*dtoresponse.ExampleResponse, len(result))
	for i, entity := range result {
		exampleResponses[i] = &dtoresponse.ExampleResponse{
			ID:      entity.ID,
			Code:    entity.Code,
			Example: entity.Example,
		}
	}

	paginatedResponse := pagination.NewPaginatedResponse(exampleResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse, response.WithMessage("Examples fetched successfully"))
}

func (h *Example) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		var clientError *utils.ClientError
		if errors.As(err, &clientError) {
			return response.BadRequest(ctx, clientError.Message, nil)
		}

		logger.Error(ctx.UserContext(), err)
		return response.InternalServerError(ctx, constants.MsgInternalServerError)
	}

	exampleResponse := &dtoresponse.ExampleResponse{
		ID:      entity.ID,
		Code:    entity.Code,
		Example: entity.Example,
	}

	return response.Success(ctx, exampleResponse, response.WithMessage("Example fetched successfully"))
}
