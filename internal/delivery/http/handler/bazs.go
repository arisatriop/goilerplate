package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/bazs"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Bazs struct {
	Validator *validator.Validate
	Usecase   bazs.Usecase
}

func NewBazs(validator *validator.Validate, usecase bazs.Usecase) *Bazs {
	return &Bazs{
		Validator: validator,
		Usecase:   usecase,
	}
}

// @Summary      Create bazs
// @Tags         bazss
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.BazsCreateRequest  true  "Bazs data"
// @Success      201      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bazss [post]
func (h *Bazs) Create(ctx *fiber.Ctx) error {
	var req dtorequest.BazsCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bazs.Bazs{
		Code: req.Code,
		Name: req.Name,
	}

	_, err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil, response.WithMessage(bazs.MsgBazsCreatedSuccessfully))
}

// @Summary      Update bazs
// @Tags         bazss
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Bazs ID"
// @Param        request  body      dtorequest.BazsUpdateRequest  true  "Bazs data"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      404      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bazss/{id} [put]
func (h *Bazs) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var req dtorequest.BazsUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bazs.Bazs{
		ID:   id,
		Code: req.Code,
		Name: req.Name,
	}

	_, err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(bazs.MsgBazsUpdatedSuccessfully))
}

// @Summary      Delete bazs
// @Tags         bazss
// @Produce      json
// @Param        id   path      string  true  "Bazs ID"
// @Success      204  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bazss/{id} [delete]
func (h *Bazs) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity := &bazs.Bazs{
		ID: id,
	}

	err := h.Usecase.Delete(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

// @Summary      List bazss
// @Tags         bazss
// @Produce      json
// @Param        keyword  query     string  false  "Search keyword"
// @Param        page     query     int     false  "Page number"   default(1)
// @Param        limit    query     int     false  "Page size"     default(10)
// @Success      200      {object}  response.PaginatedResponse{data=[]dtoresponse.BazsResponse}
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bazss [get]
func (h *Bazs) List(ctx *fiber.Ctx) error {
	var req dtorequest.BazsListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	filter := request.ToBazsFilter(&req, ctx)

	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	bazsResponses := presenter.ToBazsListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(bazsResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse, response.WithMessage(bazs.MsgBazsListFetchSuccessfully))
}

// @Summary      Get bazs by ID
// @Tags         bazss
// @Produce      json
// @Param        id   path      string  true  "Bazs ID"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BazsResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bazss/{id} [get]
func (h *Bazs) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	bazsResponse := presenter.ToBazsResponse(entity)

	return response.Success(ctx, bazsResponse, response.WithMessage(bazs.MsgBazsFetchedSuccessfully))
}
