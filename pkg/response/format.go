package response

import (
	"net/http"
	"reflect"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"

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
	// Code is set on every error: a stable, machine-readable, snake_case identifier. Clients
	// branch on it (and on the status); Message is prose for humans and may change.
	Code   string `json:"code,omitempty"`
	Data   any    `json:"data,omitempty"`
	Meta   *Meta  `json:"meta,omitempty"`
	Errors any    `json:"errors,omitempty"`
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

// CreatedAt sends 201 with a Location header naming the new resource, as RFC 9110 §15.3.2
// expects of a 201. location may be relative; clients resolve it against the request URL.
func CreatedAt(ctx *fiber.Ctx, location string, data any, options ...ResponseOption) error {
	ctx.Location(location)
	return Created(ctx, data, options...)
}

// Created sends a successful creation response. Prefer CreatedAt whenever the new resource has
// a URL of its own.
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

// Generic error codes, used when the error has no more specific one. A domain error brings its
// own (for example "bar_not_found") through apperr.
const (
	CodeBadRequest         = "bad_request"
	CodeValidationFailed   = "validation_failed"
	CodeUnauthorized       = "unauthorized"
	CodeForbidden          = "forbidden"
	CodeNotFound           = "not_found"
	CodeMethodNotAllowed   = "method_not_allowed"
	CodeConflict           = "conflict"
	CodePayloadTooLarge    = "payload_too_large"
	CodeUnprocessable      = "unprocessable_entity"
	CodeRateLimited        = "rate_limited"
	CodeInternal           = "internal_error"
	CodeServiceUnavailable = "service_unavailable"
)

// statusByKind is the one place a client error's category becomes an HTTP status.
var statusByKind = map[apperr.Kind]int{
	apperr.Invalid:         http.StatusBadRequest,
	apperr.Unauthenticated: http.StatusUnauthorized,
	apperr.Forbidden:       http.StatusForbidden,
	apperr.NotFound:        http.StatusNotFound,
	apperr.Conflict:        http.StatusConflict,
}

// Fail sends an error envelope. The helpers below are shorthands for it with the generic code
// and default message of their status.
func Fail(ctx *fiber.Ctx, status int, code, message string, details any) error {
	return ctx.Status(status).JSON(&BaseResponse{
		Success: false,
		Code:    code,
		Message: message,
		Errors:  details,
	})
}

// FailStatus sends an error envelope for a status that did not come from a domain error — a
// router 404, a 405, a body over the limit — with that status's generic code and text.
func FailStatus(ctx *fiber.Ctx, status int, message string) error {
	if message == "" {
		message = http.StatusText(status)
	}
	return Fail(ctx, status, codeForStatus(status), message, nil)
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusMethodNotAllowed:
		return CodeMethodNotAllowed
	case http.StatusConflict:
		return CodeConflict
	case http.StatusRequestEntityTooLarge:
		return CodePayloadTooLarge
	case http.StatusUnprocessableEntity:
		return CodeUnprocessable
	case http.StatusTooManyRequests:
		return CodeRateLimited
	case http.StatusServiceUnavailable:
		return CodeServiceUnavailable
	}
	if status >= http.StatusInternalServerError {
		return CodeInternal
	}
	return CodeBadRequest
}

func orDefault(message, fallback string) string {
	if message == "" {
		return fallback
	}
	return message
}

// BadRequest sends a 400 for a request that could not be parsed.
func BadRequest(ctx *fiber.Ctx, message string, details any) error {
	return Fail(ctx, http.StatusBadRequest, CodeBadRequest, orDefault(message, "Bad request"), details)
}

// ValidationError sends a 400 listing the fields that failed validation.
func ValidationError(ctx *fiber.Ctx, details any) error {
	return Fail(ctx, http.StatusBadRequest, CodeValidationFailed, "Validation failed", details)
}

// Unauthorized sends a 401.
func Unauthorized(ctx *fiber.Ctx, message string) error {
	return Fail(ctx, http.StatusUnauthorized, CodeUnauthorized, orDefault(message, constants.MsgUnauthorized), nil)
}

// Forbidden sends a 403.
func Forbidden(ctx *fiber.Ctx, message string) error {
	return Fail(ctx, http.StatusForbidden, CodeForbidden, orDefault(message, constants.MsgForbidden), nil)
}

// NotFound sends a 404.
func NotFound(ctx *fiber.Ctx, message string) error {
	return Fail(ctx, http.StatusNotFound, CodeNotFound, orDefault(message, constants.MsgResourceNotFound), nil)
}

// Conflict sends a 409.
func Conflict(ctx *fiber.Ctx, message string, details any) error {
	return Fail(ctx, http.StatusConflict, CodeConflict, orDefault(message, "Conflict"), details)
}

// UnprocessableEntity sends a 422.
func UnprocessableEntity(ctx *fiber.Ctx, message string, details any) error {
	return Fail(ctx, http.StatusUnprocessableEntity, CodeUnprocessable, orDefault(message, "Unprocessable entity"), details)
}

// TooManyRequests sends a 429. The limiter has already set Retry-After.
func TooManyRequests(ctx *fiber.Ctx, message string) error {
	return Fail(ctx, http.StatusTooManyRequests, CodeRateLimited,
		orDefault(message, "Too many requests, please try again later"), nil)
}

// InternalServerError sends a 500 with a generic message; the detail belongs in the log.
func InternalServerError(ctx *fiber.Ctx, message string) error {
	return Fail(ctx, http.StatusInternalServerError, CodeInternal, orDefault(message, constants.MsgInternalServerError), nil)
}

// HandleError answers with a use case's error. A client error (apperr) keeps its status, code
// and message; anything else is a server fault — logged with its full chain, and answered with
// a generic 500 so no internal detail reaches the caller.
func HandleError(ctx *fiber.Ctx, err error) error {
	if appErr, ok := apperr.As(err); ok {
		status, known := statusByKind[appErr.Kind]
		if known {
			return Fail(ctx, status, appErr.Code, appErr.Message, nil)
		}
	}

	logger.Error(ctx.UserContext(), err)
	return InternalServerError(ctx, "")
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
