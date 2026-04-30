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

// @Summary      Create bar
// @Tags         bars
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.BarCreateRequest  true  "Bar data"
// @Success      201      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars [post]
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

	_, err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil, response.WithMessage(bar.MsgBarCreatedSuccessfully))
}

// @Summary      Update bar
// @Tags         bars
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Bar ID"
// @Param        request  body      dtorequest.BarUpdateRequest  true  "Bar data"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      404      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [put]
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

	_, err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(bar.MsgBarUpdatedSuccessfully))
}

// @Summary      Delete bar
// @Tags         bars
// @Produce      json
// @Param        id   path      string  true  "Bar ID"
// @Success      204  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [delete]
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

// @Summary      List bars
// @Tags         bars
// @Produce      json
// @Param        keyword  query     string  false  "Search keyword"
// @Param        page     query     int     false  "Page number"   default(1)
// @Param        limit    query     int     false  "Page size"     default(10)
// @Success      200      {object}  response.PaginatedResponse{data=[]dtoresponse.BarResponse}
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars [get]
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

// @Summary      Get bar by ID
// @Tags         bars
// @Produce      json
// @Param        id   path      string  true  "Bar ID"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [get]
func (h *Bar) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	barResponse := presenter.ToBarResponse(entity)

	return response.Success(ctx, barResponse, response.WithMessage(bar.MsgBarFetchedSuccessfully))
}
