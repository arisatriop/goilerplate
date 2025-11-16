package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/delivery/http/mapper"
	"goilerplate/internal/domain/foo"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	fooApp "goilerplate/internal/application/foo"
)

type Foo struct {
	Validator           *validator.Validate
	AppplicationService fooApp.ApplicationService
	Usecase             foo.Usecase
}

func NewFoo(validator *validator.Validate, applicationService fooApp.ApplicationService, usecase foo.Usecase) *Foo {
	return &Foo{
		Validator:           validator,
		AppplicationService: applicationService,
		Usecase:             usecase,
	}
}

func (h *Foo) Create(ctx *fiber.Ctx) error {
	var req dtorequest.FooCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	foo := &foo.Foo{
		Foo: req.Foo,
	}

	if err := h.Usecase.Create(ctx.Context(), foo); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(constants.MsgSuccess))
}

func (h *Foo) Update(ctx *fiber.Ctx) error {
	var req dtorequest.FooUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	foo := &foo.Foo{
		Foo: req.Foo,
	}

	if err := h.Usecase.Update(ctx.Context(), foo); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(constants.MsgSuccess))
}

func (h *Foo) List(ctx *fiber.Ctx) error {
	filter := mapper.ToFooFilter(ctx, true)

	foos, total, err := h.Usecase.GetList(ctx.Context(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	fooResp := make([]dtoresponse.FooResponse, len(foos))
	for i, f := range foos {
		fooResp[i] = mapper.ToFooResponse(&f)
	}

	paginatedResp := pagination.NewPaginatedResponse(fooResp, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResp, response.WithMessage(constants.MsgSuccess))
}

func (h *Foo) GetByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidID, nil)
	}

	foo, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	fooResp := mapper.ToFooResponse(foo)

	return response.Success(ctx, fooResp, response.WithMessage(constants.MsgSuccess))
}

func (h *Foo) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidID, nil)
	}

	if err := h.Usecase.SoftDelete(ctx.UserContext(), id); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}
