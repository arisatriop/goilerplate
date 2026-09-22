package response

import (
	"errors"
	"net/http"
	"reflect"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Meta contains metadata about the response.
//
// Every key the API emits is camelCase. This struct used to emit request_id while the DTOs
// around it emitted accessToken and refreshTokenExpiresAt, so a single response body carried
// both conventions. See TestEnvelope_EveryKeyIsCamelCase, which fails if that returns.
//
// Pagination is embedded by pointer, so its fields are flattened into meta on a list response
// and absent everywhere else. Embedding rather than nesting keeps false and 0 meaningful:
// with omitempty on each field, "total": 0 on an empty page would vanish.
type Meta struct {
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`

	*Pagination
}

// Pagination describes the page a list response returned. It lives in meta so that data stays
// a plain array — a client can hand data straight to whatever renders the list, and read the
// page counters only if it draws a pager.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// NewPagination derives the page counters from the total and the page that was asked for.
func NewPagination(total int64, page, limit int) *Pagination {
	if limit < 1 {
		limit = 1
	}
	if page < 1 {
		page = 1
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1 && totalPages > 0,
	}
}

// BaseResponse is the standard structure for all API responses
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// ResponseOption allows customizing the response
type ResponseOption func(*BaseResponse)

// WithMeta adds metadata to the response
func WithMeta(meta *Meta) ResponseOption {
	return func(r *BaseResponse) {
		r.Meta = meta
	}
}

// WithMessage sets a custom message
func WithMessage(message string) ResponseOption {
	return func(r *BaseResponse) {
		r.Message = message
	}
}

// Success sends a successful response
func Success(ctx *fiber.Ctx, data any, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgSuccess,
		Data:    data,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusOK).JSON(response)
}

// Created sends a successful creation response
func Created(ctx *fiber.Ctx, data any, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgResourceCreatedSuccessfully,
		Data:    data,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusCreated).JSON(response)
}

// NoContent sends a successful response with no content
func NoContent(ctx *fiber.Ctx, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgOperationCompletedSuccessfully,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusNoContent).JSON(response)
}

// BadRequest sends a bad request error response
func BadRequest(ctx *fiber.Ctx, message string, errors any) error {
	return ctx.Status(http.StatusBadRequest).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgUnauthorized
	}
	return ctx.Status(http.StatusUnauthorized).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// Forbidden sends a forbidden error response
func Forbidden(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgForbidden
	}
	return ctx.Status(http.StatusForbidden).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// NotFound sends a not found error response
func NotFound(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgResourceNotFound
	}
	return ctx.Status(http.StatusNotFound).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// Conflict sends a conflict error response
func Conflict(ctx *fiber.Ctx, message string, errors any) error {
	return ctx.Status(http.StatusConflict).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// UnprocessableEntity sends an unprocessable entity error response
func UnprocessableEntity(ctx *fiber.Ctx, message string, errors any) error {
	return ctx.Status(http.StatusUnprocessableEntity).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgInternalServerError
	}
	return ctx.Status(http.StatusInternalServerError).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// ValidationError formats validation errors in a standardized way
func ValidationError(ctx *fiber.Ctx, errors any) error {
	return ctx.Status(http.StatusBadRequest).JSON(&BaseResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// Paginated sends a list response: data is the plain array, and the page counters go in meta.
//
// There used to be three descriptions of this shape and no two agreed — one in the conventions
// doc that nothing implemented, one here with a top-level "pagination" key that had no callers,
// and one in pkg/pagination that nested the array under data.items and was what the reference
// handler actually returned. This is the only one now.
//
// data is always rendered, as [] rather than null when the page is empty, so a client can
// iterate it without a nil check.
func Paginated(ctx *fiber.Ctx, data any, page *Pagination, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgSuccess,
		Data:    emptySliceIfNil(data),
		Meta:    &Meta{Pagination: page},
	}

	for _, opt := range options {
		opt(response)
	}

	// An option may have replaced Meta wholesale; the page counters belong on whatever it left.
	if response.Meta == nil {
		response.Meta = &Meta{}
	}
	response.Meta.Pagination = page

	return ctx.Status(http.StatusOK).JSON(response)
}

// TooManyRequests sends a 429 rate limit exceeded response
func TooManyRequests(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = "Too many requests, please try again later"
	}
	return ctx.Status(http.StatusTooManyRequests).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// CustomError sends a custom error response with specified status code
func CustomError(ctx *fiber.Ctx, statusCode int, message string, errors any) error {
	return ctx.Status(statusCode).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// HandleError handles errors from use case calls with consistent error responses
// It distinguishes between client errors (validation, business logic) and internal errors
// This is a reusable helper for all handlers to maintain consistent error handling
func HandleError(ctx *fiber.Ctx, err error) error {
	var clientError *utils.ClientError
	if errors.As(err, &clientError) {
		return CustomError(ctx, clientError.Code, clientError.Message, nil)
	}

	logger.Error(ctx.UserContext(), err)
	return InternalServerError(ctx, constants.MsgInternalServerError)
}

// emptySliceIfNil turns a nil slice into an empty one of the same type, so an empty page
// marshals as [] rather than null. A client iterating data should not need a nil check to
// tell "no results" from "no field".
func emptySliceIfNil(data any) any {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice && value.IsNil() {
		return reflect.MakeSlice(value.Type(), 0, 0).Interface()
	}
	return data
}
