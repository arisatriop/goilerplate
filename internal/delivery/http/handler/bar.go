package handler

import (
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/bar"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"
	"strings"

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
// @ID           barsCreate
// @Description  code must start with EXP (400 bar_code_invalid) and be unique (409 bar_code_already_exists). Retrying with the same Idempotency-Key replays the first response with Idempotency-Replayed: true.
// @Tags         bars
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header  string  true  "Unique per logical request, at most 255 characters"
// @Param        request  body      dtorequest.BarCreateRequest  true  "Bar data"
// @Success      201      {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Header       201      {string}  Location  "Path of the created bar"
// @Failure      400      {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401      {object}  response.BaseResponse
// @Failure      403      {object}  response.BaseResponse  "The caller lacks the bar create permission"
// @Failure      409      {object}  response.BaseResponse  "bar_code_already_exists, or a request with this Idempotency-Key is still being processed"
// @Failure      422      {object}  response.BaseResponse  "Idempotency-Key reused with a different payload"
// @Failure      429      {object}  response.BaseResponse
// @Header       429      {integer}  Retry-After  "Seconds until the limit resets"
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
		Code: req.Code,
		Bar:  req.Bar,
	}

	created, err := h.Usecase.Create(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	// Built from the request path rather than a fixed prefix, because the same handler is
	// mounted under /api/v1, /internal and /partner/v1.
	location := strings.TrimSuffix(ctx.Path(), "/") + "/" + created.ID
	return response.CreatedAt(ctx, location, presenter.ToBarResponse(created),
		response.WithMessage(bar.MsgBarCreatedSuccessfully))
}

// @Summary      Update bar
// @ID           barsUpdate
// @Tags         bars
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Bar ID (UUID)"
// @Description  Replaces the bar's content. code may be omitted or repeated but not changed: it is the business key (400 bar_code_immutable).
// @Param        request  body      dtorequest.BarUpdateRequest  true  "Bar data"
// @Success      200      {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      400      {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401      {object}  response.BaseResponse
// @Failure      403      {object}  response.BaseResponse  "The caller lacks the bar update permission"
// @Failure      404      {object}  response.BaseResponse
// @Failure      429      {object}  response.BaseResponse
// @Header       429      {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [put]
func (h *Bar) Update(ctx *fiber.Ctx) error {
	var param dtorequest.BarIDParam
	if err := ctx.ParamsParser(&param); err != nil {
		return response.BadRequest(ctx, "Invalid path parameters", nil)
	}
	if err := h.Validator.Struct(&param); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	var req dtorequest.BarUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	entity := &bar.Bar{
		ID:   param.ID,
		Code: req.Code,
		Bar:  req.Bar,
	}

	updated, err := h.Usecase.Update(ctx.UserContext(), entity)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, presenter.ToBarResponse(updated), response.WithMessage(bar.MsgBarUpdatedSuccessfully))
}

// @Summary      Delete bar
// @ID           barsDelete
// @Description  Soft deletes the bar. Repeating the call is 404.
// @Tags         bars
// @Produce      json
// @Param        id   path      string  true  "Bar ID (UUID)"
// @Success      200  {object}  response.BaseResponse
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}  "Malformed bar ID"
// @Failure      401  {object}  response.BaseResponse
// @Failure      403      {object}  response.BaseResponse  "The caller lacks the bar delete permission"
// @Failure      404  {object}  response.BaseResponse
// @Failure      429      {object}  response.BaseResponse
// @Header       429      {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [delete]
func (h *Bar) Delete(ctx *fiber.Ctx) error {
	var param dtorequest.BarIDParam
	if err := ctx.ParamsParser(&param); err != nil {
		return response.BadRequest(ctx, "Invalid path parameters", nil)
	}
	if err := h.Validator.Struct(&param); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	entity := &bar.Bar{
		ID: param.ID,
	}

	if err := h.Usecase.Delete(ctx.UserContext(), entity); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(bar.MsgBarDeletedSuccessfully))
}

// @Summary      List bars
// @ID           barsList
// @Description  Paginated: data is the array, and the page counters are in meta. An out-of-range page or limit is clamped, not rejected.
// @Tags         bars
// @Produce      json
// @Param        keyword  query     string  false  "Search keyword"
// @Param        page     query     int     false  "Page number"   default(1)
// @Param        limit    query     int     false  "Page size, capped at 100"  default(10)
// @Success      200      {object}  response.BaseResponse{data=[]dtoresponse.BarResponse,meta=response.Meta}
// @Failure      401      {object}  response.BaseResponse
// @Failure      403      {object}  response.BaseResponse  "The caller lacks the bar list permission"
// @Failure      429      {object}  response.BaseResponse
// @Header       429      {integer}  Retry-After  "Seconds until the limit resets"
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
	page := response.NewPagination(total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Paginated(ctx, barResponses, page, response.WithMessage(bar.MsgBarListFetchSuccessfully))
}

// @Summary      Get bar by ID
// @ID           barsGet
// @Description  A bar that does not exist is 404.
// @Tags         bars
// @Produce      json
// @Param        id   path      string  true  "Bar ID (UUID)"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}  "Malformed bar ID"
// @Failure      401  {object}  response.BaseResponse
// @Failure      403      {object}  response.BaseResponse  "The caller lacks the bar get permission"
// @Failure      404  {object}  response.BaseResponse
// @Failure      429      {object}  response.BaseResponse
// @Header       429      {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/bars/{id} [get]
func (h *Bar) Get(ctx *fiber.Ctx) error {
	var param dtorequest.BarIDParam
	if err := ctx.ParamsParser(&param); err != nil {
		return response.BadRequest(ctx, "Invalid path parameters", nil)
	}
	if err := h.Validator.Struct(&param); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	entity, err := h.Usecase.GetByID(ctx.UserContext(), param.ID)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	barResponse := presenter.ToBarResponse(entity)

	return response.Success(ctx, barResponse, response.WithMessage(bar.MsgBarFetchedSuccessfully))
}
