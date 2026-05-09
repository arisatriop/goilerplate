package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/bas"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
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

// @Summary      Create bas
// @Tags         bas
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.BasCreateRequest  true  "Bas data"
// @Success      201      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bas [post]
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
		Name: req.Name,
	}

	result, err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, presenter.ToBasResponse(result), response.WithMessage(bas.MsgBasCreatedSuccessfully))
}

// @Summary      Update bas
// @Tags         bas
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Bas ID"
// @Param        request  body      dtorequest.BasUpdateRequest  true  "Bas data"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      404      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bas/{id} [put]
func (h *Bas) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var req dtorequest.BasUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bas.Bas{
		ID:   id,
		Code: req.Code,
		Name: req.Name,
	}

	result, err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, presenter.ToBasResponse(result), response.WithMessage(bas.MsgBasUpdatedSuccessfully))
}

// @Summary      Delete bas
// @Tags         bas
// @Produce      json
// @Param        id   path      string  true  "Bas ID"
// @Success      204  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bas/{id} [delete]
func (h *Bas) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity := &bas.Bas{
		ID: id,
	}

	if err := h.Usecase.Delete(ctx.UserContext(), entity); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

// @Summary      List bas
// @Tags         bas
// @Produce      json
// @Param        keyword  query     string  false  "Search keyword"
// @Param        page     query     int     false  "Page number"   default(1)
// @Param        limit    query     int     false  "Page size"     default(10)
// @Success      200      {object}  response.PaginatedResponse{data=[]dtoresponse.BasResponse}
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bas [get]
func (h *Bas) List(ctx *fiber.Ctx) error {
	var req dtorequest.BasListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	filter := request.ToBasFilter(&req, ctx)

	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	basResponses := presenter.ToBasListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(basResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse, response.WithMessage(bas.MsgBasListFetchSuccessfully))
}

// @Summary      Get bas by ID
// @Tags         bas
// @Produce      json
// @Param        id   path      string  true  "Bas ID"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BasResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bas/{id} [get]
func (h *Bas) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	entity, err := h.Usecase.GetByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, presenter.ToBasResponse(entity), response.WithMessage(bas.MsgBasFetchedSuccessfully))
}
